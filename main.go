package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
	. "github.com/cyoung/rpi"
)

/*
 * Raspberry Pi 4 GPIO
 *
 *         3v3 Power   01 02   5v Power
 *      GPIO 2 (SDA)   03 04   5v Power
 *      GPIO 3 (SCL)   05 06   Ground
 *   GPIO 4 (GPCLK0)   07 08   GPIO 14 (TXD)
 *            Ground   09 10   GPIO 15 (RXD)
 *           GPIO 17   11 12   GPIO 18 (PCM_CLK)
 *           GPIO 27   13 14   Ground
 *           GPIO 22   15 16   GPIO 23
 *         3v3 Power   17 18   GPIO 24
 *    GPIO 10 (MOSI)   19 20   Ground
 *     GPIO 9 (MISO)   21 22   GPIO 25
 *    GPIO 11 (SCLK)   23 24   GPIO 8 (CE0)
 *            Ground   25 26   GPIO 7 (CE1)
 *    GPIO 0 (ID_SD)   27 28   GPIO 1 (ID_SC)
 *            GPIO 5   29 30   Ground
 *            GPIO 6   31 32   GPIO 12 (PWM0)
 *    GPIO 13 (PWM1)   33 34   Ground
 *  GPIO 19 (PCM_FS)   35 36   GPIO 16
 *           GPIO 26   37 38   GPIO 20 (PCM_DIN)
 *            Ground   39 40   GPIO 21 (PCM_DOUT)
 */

var powerState bool

const (
	powerPin = 7
	lightPin = 1
)

type IntensityResponse struct {
	Value int `json:"Value"`
}

type PowerRequest struct {
	On bool `json:"On"`
}

type PowerResponse struct {
	On bool `json:"On"`
}

func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to Pi Lights!")
}

func handleIntensity(w http.ResponseWriter, r *http.Request) {
	PinMode(lightPin, INPUT)
	res := DigitalRead(lightPin)

	var intensity int
	if res == LOW {
		intensity = 1
	} else {
		intensity = 0
	}

	fmt.Printf("Intensity: %d\n", intensity)

	json.NewEncoder(w).Encode(IntensityResponse{
		Value: intensity,
	})
}

func handlePower(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		fmt.Printf("Power: %t\n", powerState)

		json.NewEncoder(w).Encode(PowerResponse{
			On: powerState,
		})
	} else if r.Method == http.MethodPost {
		reqBody, _ := ioutil.ReadAll(r.Body)
		var power PowerRequest
		json.Unmarshal(reqBody, &power)

		setPower(power.On)
		w.WriteHeader(http.StatusOK)
	}
}

func setPower(on bool) {
	fmt.Printf("Setting Power: %t\n", on)
	powerState = on

	PinMode(powerPin, OUTPUT)
	if on {
		DigitalWrite(powerPin, HIGH)
	} else {
		DigitalWrite(powerPin, LOW)
	}
}

func handleRequests(port int) {
	myRouter := mux.NewRouter().StrictSlash(true)
	myRouter.HandleFunc("/", homePage)
	myRouter.HandleFunc("/light/intensity", handleIntensity)
	myRouter.HandleFunc("/light/power", handlePower).Methods(http.MethodGet, http.MethodPost)

	fmt.Printf("server starting on port :%d\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), myRouter))
}

func main() {
	if len(os.Args) > 2 {
		fmt.Println("Usage:", os.Args[0], "[port:10000]")
		return
	}

	var port int
	if len(os.Args) == 2 {
		port, _ = strconv.Atoi(os.Args[1])
	} else {
		port = 10000
	}

	WiringPiSetup()
	setPower(false)

	handleRequests(port)
}
