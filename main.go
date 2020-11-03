package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
)

type config struct {
	cw20ContractAddr string
	keyName          string
	chainID          string
	rpc              string
	cliHome          string
	pathToWasmCLI    string
	defaultAmount    int
	debug            bool
	ip               string
	port             int
}

var conf = config{
	cw20ContractAddr: "cosmos1kfz3mj84atqjld0ge9eccujvqqkqdr4qqs9ud7",
	defaultAmount:    100,
	chainID:          "hackatom-wasm",
	cliHome:          "/home/ec2-user/faucet-home",
	pathToWasmCLI:    "/home/ec2-user/go/bin/wasmcli",
	rpc:              "https://rpc.heldernet.cosmwasm.com:443",
	debug:            true,
	ip:               "0.0.0.0",
	port:             8080,
	keyName:          "token",
}

// FaucetReqInfo is the body that needs to be sent to /faucet
type FaucetReqInfo struct {
	Address string `json:"address"`
}

func main() {
	printInfo()

	http.HandleFunc("/faucet", faucet)
	http.ListenAndServe(fmt.Sprintf("%s:%d", conf.ip, conf.port), nil)
}

func printInfo() {
	log.Println("Starting CW20 faucet...")

	fmt.Println()
	log.Println("Chain info")
	log.Printf("        RPC:                   %s\n", conf.rpc)
	log.Printf("        CW20 Contract Address: %s\n", conf.cw20ContractAddr)
	log.Printf("        Chain ID:              %s\n", conf.chainID)

	fmt.Println()
	log.Println("Configuration")
	log.Printf("        Default Amount: %d\n", conf.defaultAmount)
	log.Printf("        CLI Home:       %s\n", conf.cliHome)
	log.Printf("        Key Name:       %s\n", conf.keyName)
	log.Printf("        Debug:          %v\n", conf.debug)

	fmt.Println()
	log.Printf("Server is listening on %s:%d\n", conf.ip, conf.port)
}

// handler for the faucet endpoint
func faucet(res http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {

	}
	defer req.Body.Close()
	faucetReqInfo := FaucetReqInfo{}
	err = json.Unmarshal(body, &faucetReqInfo)
	if err != nil {
		if conf.debug {
			log.Printf("ERROR: Couldn't unmarshal request body: %s\n", err.Error())
		}

		res.WriteHeader(400)
		res.Write([]byte(err.Error()))
		return
	}

	err = transferCoins(conf.defaultAmount, faucetReqInfo.Address)
	if err != nil {
		if conf.debug {
			log.Printf("ERROR: Transfer of coins to address %s failed: %s\n", faucetReqInfo.Address, err.Error())
		}
		res.WriteHeader(500)
		res.Write([]byte(err.Error()))
		return
	}

	log.Printf("INFO: Address %s reuested funds\n", faucetReqInfo.Address)
	res.Header().Set("Access-Control-Allow-Origin", "*")
	res.WriteHeader(200)
	res.Write([]byte{})
}

// executing transfer with the help of the cli
func transferCoins(amount int, address string) error {
	/*{
		"transfer": {
			"amount": "",      // numeric value as string ("529")
			"recipient": ""    // account address
		}
	}*/
	transferMsg := fmt.Sprintf("{\"transfer\":{\"amount\":\"%d\",\"recipient\":\"%s\"}}", conf.defaultAmount, address)

	// wasmcli tx wasm execute contractAddr msg --node node --chain-id cid --keyring-backend test --from keyName --yes --home home
	transferCmd := exec.Command(conf.pathToWasmCLI,
		"tx", "wasm", "execute", conf.cw20ContractAddr, transferMsg,
		"--node", conf.rpc,
		"--chain-id", conf.chainID,
		"--keyring-backend", "test",
		"--from", conf.keyName,
		"--home", conf.cliHome,
		"--gas-prices", "0.025ucosm",
		"--gas", "auto",
		"--gas-adjustment", "1.2",
		"--yes",
	)

	bz, err := transferCmd.CombinedOutput()
	log.Println(string(bz))
	if err != nil {
		return err
	}

	//return transferCmd.Run()
	return nil
}
