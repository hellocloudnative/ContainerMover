package master

import (
	"ContainerMover/pkg/logger"
	"ContainerMover/pkg/sshcmd/sshutil"
	"fmt"
	"io/ioutil"
	"math/big"
	"net"
	"os"
	"sigs.k8s.io/yaml"
	"strings"
)

type ContainerMoverConfig struct {
	Nodes []string
	//SSHConfig
	User       string
	Passwd     string
	PrivateKey string
	PkPassword string
}

const (
	ErrorExitOSCase = -1                    // 错误直接退出类型
	ErrorNodesEmpty = "your Node is empty." // node节点ip为空
)

var (
	Images        bool
	ImageNames    []string
	ImageListFile string
	SrcType       string
	DstType       string
	Namespace     string
	AllImages     bool
	Nodes         []string
	SSHConfig     sshutil.SSH
	message       string
)

const defaultConfigPath = "/.containerMover"
const defaultConfigFile = "/config.yaml"

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
func Cmp(a, b net.IP) int {
	aa := ipToInt(a)
	bb := ipToInt(b)

	if aa == nil || bb == nil {
		logger.Error("ip range %s-%s is invalid", a.String(), b.String())
		os.Exit(-1)
	}
	return aa.Cmp(bb)
}
func ipToInt(ip net.IP) *big.Int {
	if v := ip.To4(); v != nil {
		return big.NewInt(0).SetBytes(v)
	}
	return big.NewInt(0).SetBytes(ip.To16())
}
func stringToIP(i string) net.IP {
	return net.ParseIP(i).To4()
}

// NextIP returns IP incremented by 1
func NextIP(ip net.IP) net.IP {
	i := ipToInt(ip)
	return intToIP(i.Add(i, big.NewInt(1)))
}
func intToIP(i *big.Int) net.IP {
	return net.IP(i.Bytes())
}

// Load is
func (c *ContainerMoverConfig) Load(path string) (err error) {
	if path == "" {
		home, _ := os.UserHomeDir()
		path = home + defaultConfigPath + defaultConfigFile
	}

	y, err := ioutil.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read config file %s failed %w", path, err)
	}

	err = yaml.Unmarshal(y, c)
	if err != nil {
		return fmt.Errorf("unmarshal config file failed: %w", err)
	}

	Nodes = c.Nodes
	SSHConfig.User = c.User
	SSHConfig.Password = c.Passwd
	SSHConfig.PkFile = c.PrivateKey
	SSHConfig.PkPassword = c.PkPassword
	return
}

// Dump is
func (c *ContainerMoverConfig) Dump(path string) {
	home, _ := os.UserHomeDir()
	if path == "" {
		path = home + defaultConfigPath + defaultConfigFile
	}
	Nodes = ParseIPs(Nodes)
	c.Nodes = ParseIPs(Nodes)
	c.User = SSHConfig.User
	c.Passwd = SSHConfig.Password
	c.PrivateKey = SSHConfig.PkFile
	c.PkPassword = SSHConfig.PkPassword

	y, err := yaml.Marshal(c)
	if err != nil {
		logger.Error("dump config file failed: %s", err)
	}

	err = os.MkdirAll(home+defaultConfigPath, os.ModePerm)
	if err != nil {
		logger.Warn("create default kubeprince  config dir failed, please create it by your self mkdir -p /root/.kubeprince && touch /root/.kubeprince /config.yaml")
	}

	if err = ioutil.WriteFile(path, y, 0644); err != nil {
		logger.Warn("write to file %s failed: %s", path, err)
	}
}

func ExitInitCase() bool {
	// 重大错误直接退出, 不保存配置文件
	if len(Nodes) == 0 {
		message = ErrorNodesEmpty
	}
	// 用户不写 --passwd, 默认走pk, 秘钥如果没有配置ssh互信, 则验证ssh的时候报错. 应该属于preRun里面
	// first to auth password, second auth pk.
	// 如果初始状态都没写, 默认都为空. 报这个错
	//if SSHConfig.Password == "" && SSHConfig.PkFile == "" {
	//	message += ErrorMessageSSHConfigEmpty
	//}
	if message != "" {
		logger.Error(message + "please check your command is ok?")
		return true
	}
	return false
}
