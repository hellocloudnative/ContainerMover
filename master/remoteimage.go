package master

import (
	"context"
	"fmt"
	"github.com/docker/docker/client"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"io"
	"os"
	"path/filepath"
)

func MigrateImageRemotely(srcType, dstType, namespace, imageName string, host, username, password string) error {
	var err error
	// 连接到Docker客户端
	dockerClient, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return fmt.Errorf("failed to create Docker client: %v", err)
	}
	defer dockerClient.Close()

	switch srcType {
	case "docker":
		switch dstType {
		case "containerd":
			if srcType != "docker" || dstType != "containerd" {
				return fmt.Errorf("unsupported migration from %s to %s", srcType, dstType)
			}
			sshConfig := &ssh.ClientConfig{
				User: username,
				Auth: []ssh.AuthMethod{
					ssh.Password(password),
				},
				HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			}

			sshClient, err := ssh.Dial("tcp", host+":22", sshConfig)
			if err != nil {
				return fmt.Errorf("failed to dial SSH: %v", err)
			}
			defer sshClient.Close()

			tarPath, err := createDockerImageTar(dockerClient, imageName)
			if err != nil {
				return err
			}
			defer os.Remove(tarPath)

			sftpClient, err := sftp.NewClient(sshClient)
			if err != nil {
				return fmt.Errorf("failed to create SFTP client: %v", err)
			}
			defer sftpClient.Close()

			remotePath := filepath.Join("/tmp", filepath.Base(tarPath))
			if err := sftpUploadFile(sftpClient, tarPath, remotePath); err != nil {
				return err
			}
			if err := sshRunCommand(sshClient, fmt.Sprintf("ctr --namespace %s images import %s", namespace, remotePath)); err != nil {
				return err
			}
			if err := sshRunCommand(sshClient, fmt.Sprintf("rm %s", remotePath)); err != nil {
				return err
			}
		case "other-runtime":
			fmt.Printf("Migration to %s is not supported yet.\n", dstType)
		default:
			return fmt.Errorf("unsupported destination runtime: %s", dstType)
		}
	default:
		return fmt.Errorf("unsupported source runtime: %s", srcType)
	}
	return nil
}

// createDockerImageTar
func createDockerImageTar(dockerClient *client.Client, imageName string) (string, error) {
	img, _, err := dockerClient.ImageInspectWithRaw(context.Background(), imageName)
	if err != nil {
		return "", fmt.Errorf("failed to inspect image %s: %v", imageName, err)
	}

	tarPath := filepath.Join("/tmp", fmt.Sprintf("%s.tar", img.ID))
	tarFile, err := os.Create(tarPath)
	if err != nil {
		return "", fmt.Errorf("failed to create tar file: %v", err)
	}
	defer tarFile.Close()

	tarStream, err := dockerClient.ImageSave(context.Background(), []string{imageName})
	if err != nil {
		return "", fmt.Errorf("failed to save image %s: %v", imageName, err)
	}
	defer tarStream.Close()

	if _, err := io.Copy(tarFile, tarStream); err != nil {
		return "", fmt.Errorf("failed to write tar stream to file: %v", err)
	}

	return tarPath, nil
}

// sftpUploadFile
func sftpUploadFile(sftpClient *sftp.Client, localPath, remotePath string) error {
	localFile, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("failed to open local file %s: %v", localPath, err)
	}
	defer localFile.Close()

	remoteFile, err := sftpClient.Create(remotePath)
	if err != nil {
		return fmt.Errorf("failed to create remote file %s: %v", remotePath, err)
	}
	defer remoteFile.Close()

	if _, err := io.Copy(remoteFile, localFile); err != nil {
		return fmt.Errorf("failed to upload file: %v", err)
	}

	return nil
}

// sshRunCommand
func sshRunCommand(sshClient *ssh.Client, cmd string) error {
	session, err := sshClient.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create SSH session: %v", err)
	}
	defer session.Close()

	output, err := session.CombinedOutput(cmd)
	if err != nil {
		return fmt.Errorf("failed to run command '%s': %v, output: %s", cmd, err, output)
	}

	return nil
}
