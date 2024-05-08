package master

import (
	"ContainerMover/pkg/logger"
	"context"
	"fmt"
	"github.com/cheggaaa/pb/v3"
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
	logger.Info("Uploading image  %s to remote server...", imageName)
	// 创建进度条
	bar := pb.New(int(img.VirtualSize)).Set(pb.Bytes, true).SetWidth(80)
	bar.Start()
	progressReader := &progressReader{Reader: tarStream, bar: bar}
	if _, err := io.Copy(tarFile, progressReader); err != nil {
		return "", fmt.Errorf("failed to write tar stream to file: %v", err)
	}
	bar.Finish()
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

	// 获取文件大小
	fileInfo, err := localFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat local file %s: %v", localPath, err)
	}
	logger.Info("Importing file  to remote server...")
	// 创建进度条
	bar := pb.New64(fileInfo.Size()).Set(pb.Bytes, true).SetWidth(80)
	bar.Start()
	progressReader := &progressReader{Reader: localFile, bar: bar}

	if _, err := io.Copy(remoteFile, progressReader); err != nil {
		return fmt.Errorf("failed to upload file: %v", err)
	}
	bar.Finish()
	return nil
}

// sshRunCommand 在远程服务器上执行 SSH 命令
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
