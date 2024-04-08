package bin

import (
	"ContainerMover/pkg/sshcmd/sshutil"
	"fmt"
	"github.com/wonderivan/logger"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"sigs.k8s.io/yaml"
	"strconv"
	"strings"
	"time"
)

var (
	SSHConfig sshutil.SSH
	ApiServer string
)

func FileExist(path string) bool {
	_, err := os.Stat(path)
	return err == nil || os.IsExist(err)
}

// 返回/etc/hosts记录
func getApiserverHost(ipAddr string) (host string) {
	return fmt.Sprintf("%s %s", ipAddr, ApiServer)
}

// VersionToInt v1.15.6  => 115
func VersionToInt(version string) int {
	// v1.15.6  => 1.15.6
	version = strings.Replace(version, "v", "", -1)
	versionArr := strings.Split(version, ".")
	if len(versionArr) >= 2 {
		versionStr := versionArr[0] + versionArr[1]
		if i, err := strconv.Atoi(versionStr); err == nil {
			return i
		}
	}
	return 0
}

func GetRemoteHostName(hostIP string) string {
	hostName := SSHConfig.CmdToString(hostIP, "hostname", "")
	return strings.ToLower(hostName)
}

// 根据yaml转换kubeadm结构
func KubeadmDataFromYaml(context string) *KubeadmType {
	yamls := strings.Split(context, "---")
	if len(yamls) > 0 {
		for _, y := range yamls {
			cfg := strings.TrimSpace(y)
			if cfg == "" {
				continue
			} else {
				kubeadm := &KubeadmType{}
				if err := yaml.Unmarshal([]byte(cfg), kubeadm); err == nil {
					//
					if kubeadm.Kind == "ClusterConfiguration" {
						if kubeadm.Networking.DnsDomain == "" {
							kubeadm.Networking.DnsDomain = "cluster.local"
						}
						return kubeadm
					}
				}
			}
		}
	}
	return nil
}

type KubeadmType struct {
	Kind      string `yaml:"kind,omitempty"`
	ApiServer struct {
		CertSANs []string `yaml:"certSANs,omitempty"`
	} `yaml:"apiServer"`
	Networking struct {
		DnsDomain string `yaml:"dnsDomain,omitempty"`
	} `yaml:"networking"`
}

// 获取kubeprince绝对路径
func FetchubePrinceAbsPath() string {
	ex, _ := os.Executable()
	exPath, _ := filepath.Abs(ex)
	return exPath
}

// VersionToIntAll v1.19.1 ==> 1191
func VersionToIntAll(version string) int {
	version = strings.Replace(version, "v", "", -1)
	arr := strings.Split(version, ".")
	if len(arr) >= 3 {
		str := arr[0] + arr[1] + arr[2]
		if i, err := strconv.Atoi(str); err == nil {
			return i
		}
	}
	return 0
}

// 提取ip
func IpFormat(host string) string {
	ipAndPort := strings.Split(host, ":")
	return ipAndPort[0]
}

func AddReformat(host string) string {
	if strings.Index(host, ":") == -1 {
		host = fmt.Sprintf("%s:22", host)
	}
	return host
}

func readFile(name string) string {
	content, err := ioutil.ReadFile(name)
	if err != nil {
		logger.Error("[globals] read file err is : %s", err)
		return ""
	}

	return string(content)
}

func sshAuthMethod(password, pkFile string) ssh.AuthMethod {
	var am ssh.AuthMethod
	if password != "" {
		am = ssh.Password(password)
	} else {

		pkData := readFile(pkFile)
		pk, _ := ssh.ParsePrivateKey([]byte(pkData))
		am = ssh.PublicKeys(pk)
	}
	return am
}

// ssh connect
func Connect(user, password, pkFile, host string) (*ssh.Session, error) {
	auth := []ssh.AuthMethod{sshAuthMethod(password, pkFile)}
	config := ssh.Config{
		Ciphers: []string{"aes128-ctr", "aes192-ctr", "aes256-ctr", "aes128-gcm@openssh.com", "arcfour256"},
	}
	clientConfig := &ssh.ClientConfig{
		User:    user,
		Auth:    auth,
		Timeout: time.Duration(1) * time.Minute,
		Config:  config,
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) (err error) {
			return nil
		},
	}
	addr := AddReformat(host)
	client, err := ssh.Dial("tcp", addr, clientConfig)
	if err != nil {
		return nil, err
	}
	session, err := client.NewSession()
	if err != nil {
		return nil, err
	}
	modes := ssh.TerminalModes{
		ssh.ECHO:          0,     // disable echoing
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}

	if err := session.RequestPty("xterm", 80, 40, modes); err != nil {
		return nil, err
	}
	return session, nil
}

var message string

func ipToInt(ip net.IP) *big.Int {
	if v := ip.To4(); v != nil {
		return big.NewInt(0).SetBytes(v)
	}
	return big.NewInt(0).SetBytes(ip.To16())
}

func intToIP(i *big.Int) net.IP {
	return net.IP(i.Bytes())
}

func stringToIP(i string) net.IP {
	return net.ParseIP(i).To4()
}

// NextIP returns IP incremented by 1
func NextIP(ip net.IP) net.IP {
	i := ipToInt(ip)
	return intToIP(i.Add(i, big.NewInt(1)))
}

// Cmp compares two IPs, returning the usual ordering:
// a < b : -1
// a == b : 0
// a > b : 1
func Cmp(a, b net.IP) int {
	aa := ipToInt(a)
	bb := ipToInt(b)

	if aa == nil || bb == nil {
		logger.Error("ip range %s-%s is invalid", a.String(), b.String())
		os.Exit(-1)
	}
	return aa.Cmp(bb)
}

// ParseIPs 解析ip 192.168.0.2-192.168.0.6
func ParseIPs(ips []string) []string {
	return DecodeIPs(ips)
}

func DecodeIPs(ips []string) []string {
	var res []string
	var port string
	for _, ip := range ips {
		port = "22"
		if ipport := strings.Split(ip, ":"); len(ipport) == 2 {
			ip = ipport[0]
			port = ipport[1]
		}
		if iprange := strings.Split(ip, "-"); len(iprange) == 2 {
			for Cmp(stringToIP(iprange[0]), stringToIP(iprange[1])) <= 0 {
				res = append(res, fmt.Sprintf("%s:%s", iprange[0], port))
				iprange[0] = NextIP(stringToIP(iprange[0])).String()
			}
		} else {
			if stringToIP(ip) == nil {
				logger.Error("ip [%s] is invalid", ip)
				os.Exit(1)
			}
			res = append(res, fmt.Sprintf("%s:%s", ip, port))
		}
	}
	return res
}

// GetMajorMinorInt
func GetMajorMinorInt(version string) (major, minor int) {
	// alpha beta rc version
	if strings.Contains(version, "-") {
		v := strings.Split(version, "-")[0]
		version = v
	}
	version = strings.Replace(version, "v", "", -1)
	versionArr := strings.Split(version, ".")
	if len(versionArr) >= 2 {
		majorStr := versionArr[0] + versionArr[1]
		minorStr := versionArr[2]
		if major, err := strconv.Atoi(majorStr); err == nil {
			if minor, err := strconv.Atoi(minorStr); err == nil {
				return major, minor
			}
		}
	}
	return 0, 0
}

func CanUpgradeByNewVersion(new, old string) error {
	newMajor, newMinor := GetMajorMinorInt(new)
	major, minor := GetMajorMinorInt(old)

	// sealos change cri to containerd when version more than 1.20.0
	if newMajor == 120 && major == 119 {
		return fmt.Errorf("sealos change cri to containerd when Version greater than 1.20! New version: %s, current version: %s", new, old)
	}
	// case one:  new major version <  old major version
	// 1.18.8     1.19.1
	if newMajor < major {
		return fmt.Errorf("kubernetes new version is lower than current version! New version: %s, current version: %s", new, old)
	}
	// case two:  new major version = old major version ; new minor version <= old minor version
	// 1.18.0   1.18.1
	if newMajor == major && newMinor <= minor {
		return fmt.Errorf("kubernetes new version is lower/equal than current version! New version: %s, current version: %s", new, old)
	}

	// case three : new major version > old major version +1;
	// 1.18.2    1.16.10
	if newMajor > major+1 {
		return fmt.Errorf("kubernetes new version is bigger than current version, more than one major version is not allowed! New version: %s, current version: %s", new, old)
	}
	return nil
}

func For120(version string) bool {
	newMajor, _ := GetMajorMinorInt(version)
	// // kubernetes gt 1.20, use Containerd instead of docker
	if newMajor >= 120 {
		logger.Info("install version is: %s, Use kubeadm v1beta2 InitConfig,OCI use containerd instead", version)
		return true
	} else {
		//logger.Info("install version is: %s, Use kubeadm v1beta1 InitConfig, docker", version)
		return false
	}

}

func SliceRemoveStr(ss []string, s string) (result []string) {
	for _, v := range ss {
		if v != s {
			result = append(result, v)
		}
	}
	return
}

// 判断当前host的hostname
func isHostName(master, host string) string {
	hostString := SSHConfig.CmdToString(master, "kubectl get nodes | grep -v NAME  | awk '{print $1}'", ",")
	hostName := SSHConfig.CmdToString(host, "hostname", "")
	logger.Debug("hosts %v", hostString)
	hosts := strings.Split(hostString, ",")
	var name string
	for _, h := range hosts {
		if strings.TrimSpace(h) == "" {
			continue
		} else {
			hh := strings.ToLower(h)
			fromH := strings.ToLower(hostName)
			if hh == fromH {
				name = h
				break
			}
		}
	}
	return name
}
