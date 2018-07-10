
package main

import (
        "encoding/json"
        "fmt"
        "net/http"
	"io/ioutil"
	"os/exec"
	"time"
	"strings"
	"os"
	"net"
	"strconv"
	"sync"
	"github.com/rzetterberg/elmobd"
)

// this is the server url to fetch commands
var remoteServerURL = "http://127.0.0.1:10000/all"
var computerVoiceType = "/home/rpi3b/Desktop/voices/cmu_us_slt.flitevox"
var glob_location string // global variable for storing the last location from GPS
var glob_date string // global variable to store local time
var glob_counter int
var mutex = &sync.Mutex{}

// Article - Our struct for all articles
type Article struct {
        Id      int    `json:"Id"`
        Title   string `json:"Title"`
        Desc    string `json:"desc"`
        Content string `json:"content"`
}

func carNotification(notificationSound string, message string) {
	myCmd := "sudo play " + notificationSound + " | flite -voice " + computerVoiceType + " -v " + "\""+ message + "\""
	soundCmd := exec.Command("sudo", "bash", "-c",  myCmd)
        soundOut, err := soundCmd.Output()
        if err != nil {
		err.Error()
                //panic(err)
        }
	fmt.Println(string(soundOut))
}

func soxPlay(audioPath string) {
        voiceCmd := exec.Command("sudo", "play", audioPath)
        voiceOut, err := voiceCmd.Output()
	if err != nil {
                err.Error()
		//panic(err)
        }
	fmt.Println(string(voiceOut))
}

func algSay(phrase string) {
	fmt.Println("Saying: \"" + phrase + "\"")
	voiceCmd := exec.Command("sudo", "flite", "-voice", computerVoiceType, "-v", phrase)
        voiceOut, err := voiceCmd.Output()
	if err != nil {
		panic(err)
        }
        fmt.Println(string(voiceOut))
}

// the idea is to talk to a know server for json formated responses, given your id, it will return the command expected to be run.
func fetchCommand(myId string) Article {
	res, err := http.Get(remoteServerURL)
	if err != nil {
		panic(err.Error())
	}
	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		panic(err.Error())
	}

	var data Article
	json.Unmarshal(body, &data)
	fmt.Printf("Results: %v\n", data)
	//fmt.Printf("%v\n%v\n", data.Title, data.Content)
	return data
}

func loopCamera() {
	myCmd := fetchCommand("randomid")
	if myCmd.Title == "photonow" {
		dev, err := elmobd.NewDevice("/dev/ttyUSB0", false)

		var obd_concat string
		obd_concat = "OBD__"
		if err != nil {
			fmt.Println("Failed to create new device", err)
			return
		} else {

		// get obd data
		// NewVehicleSpeed
		speed, err := dev.RunOBDCommand(elmobd.NewVehicleSpeed())
		if err != nil {
		fmt.Println("Failed to get NewVehicleSpeed", err)
			soxPlay("/home/rpi3b/Desktop/VW_sounds/BRA_VW_SAVEIRO_INSTRUMENT_SOUNDS/VW_SVRO_DOT_BEEP.flac")			
		} else {
			//fmt.Println("NewVehicleSpeed ", speed.ValueAsLit())
			obd_concat += "SPD_" + speed.ValueAsLit()
		}
		// NewThrottlePosition
		tp, err := dev.RunOBDCommand(elmobd.NewThrottlePosition())
		if err != nil {
		fmt.Println("Failed to get NewThrottlePosition", err)
			soxPlay("/home/rpi3b/Desktop/VW_sounds/BRA_VW_SAVEIRO_INSTRUMENT_SOUNDS/VW_SVRO_DOT_BEEP.flac")			
		} else {
			//fmt.Println("NewThrottlePosition ", tp.ValueAsLit())
			obd_concat += "_TRP_" + tp.ValueAsLit()	
		}
		// NewEngineRPM
		rpm, err := dev.RunOBDCommand(elmobd.NewEngineRPM())
		if err != nil {
		fmt.Println("Failed to get NewEngineRPM", err)
			soxPlay("/home/rpi3b/Desktop/VW_sounds/BRA_VW_SAVEIRO_INSTRUMENT_SOUNDS/VW_SVRO_DOT_BEEP.flac")			
		} else {
			//fmt.Println("NewEngineRPM ", rpm.ValueAsLit())
			obd_concat += "_RPM_" + rpm.ValueAsLit()
		}
		// NewCoolantTemperature
		collanttemp, err := dev.RunOBDCommand(elmobd.NewCoolantTemperature())
		if err != nil {
		fmt.Println("Failed to get NewCoolantTemperature", err)
			soxPlay("/home/rpi3b/Desktop/VW_sounds/BRA_VW_SAVEIRO_INSTRUMENT_SOUNDS/VW_SVRO_DOT_BEEP.flac")			
		} else {
			fmt.Println("NewCoolantTemperature ", collanttemp.ValueAsLit())
			obd_concat += "_TEMP_" + collanttemp.ValueAsLit()
		}
		}		

		dateCmd := exec.Command("/home/rpi3b/Desktop/kdocarro/kdocarro_client/cam",
		"0",
		"/home/rpi3b/Desktop/kdocarro/kdocarro_client/captures/KDP_" + glob_date + glob_location + strconv.Itoa(glob_counter) + obd_concat + ".jpg")

		mutex.Lock()
		glob_counter++
		mutex.Unlock()

		dateOut, err := dateCmd.Output()
		if err != nil {
			panic(err)
		}
		fmt.Println(string(dateOut))
		go soxPlay("/home/rpi3b/Desktop/VW_sounds/BRA_VW_SAVEIRO_INSTRUMENT_SOUNDS/VW_SVRO_CLUSTER_CLICK.flac")

	}
}

/* A Simple function to verify error */
func CheckError(err error) {
    if err  != nil {
        fmt.Println("Error: " , err)
        os.Exit(0)
    }
}
 
func gpsTracker() {
    /* listen for the gps broadcaster and store the position it gives us*/   
    ServerAddr,err := net.ResolveUDPAddr("udp",":11098")
    CheckError(err)
 
    /* Now listen at selected port */
    ServerConn, err := net.ListenUDP("udp", ServerAddr)
    CheckError(err)
    defer ServerConn.Close()
 
    buf := make([]byte, 1024)
 
    for {
        n,addr,err := ServerConn.ReadFromUDP(buf)
        fmt.Println("Received ",string(buf[0:n]), " from ",addr)

	stringToSplit := string(buf[0:n])
	splited := strings.Split(stringToSplit, ";")

	receivedDate := splited[0]
	receivedCoordinates := splited[1]

	receivedDate = strings.Replace(receivedDate, "/", "-", -1)
	receivedDate = strings.Replace(receivedDate, ":", "-", -1)

	fmt.Println("Date: ", receivedDate)
	fmt.Println("Location", receivedCoordinates)

	mutex.Lock()
	glob_date = receivedDate
	glob_location = receivedCoordinates
	mutex.Unlock()

	go soxPlay("/home/rpi3b/Desktop/VW_sounds/BRA_VW_SAVEIRO_INSTRUMENT_SOUNDS/VW_SVRO_CLUSTER_BELL.flac")


        if err != nil {
            fmt.Println("Error: ",err)
        } 
    }
}


func main() {
	//carNotification("/home/rpi3b/Desktop/VW_SVRO_CLUSTER_BELL.wav", "Warning: Engine temperature is critical.")
	//algSay("Testing sound notifications.")
	//time.Sleep(500 * time.Millisecond)
	//soxPlay("/home/rpi3b/Desktop/VW_sounds/BRA_VW_SAVEIRO_INSTRUMENT_SOUNDS/VW_SVRO_CLUSTER_BELL.flac")
        
        //time.Sleep(450 * time.Millisecond)
        //for i:= 0; i < 3; i++ {
        //        soxPlay("/home/rpi3b/Desktop/VW_sounds/BRA_VW_SAVEIRO_INSTRUMENT_SOUNDS/VW_SVRO_CLUSTER_CLICK.flac")
        	
        //}
        //time.Sleep(450 * time.Millisecond)

        //for i:= 0; i < 3; i++ {
        //        soxPlay("/home/rpi3b/Desktop/VW_sounds/BRA_VW_SAVEIRO_INSTRUMENT_SOUNDS/VW_SVRO_DASH_BEEP.flac")
       	//	time.Sleep(400 * time.Millisecond)
       		
	// }
        //time.Sleep(450 * time.Millisecond)

        //for i:= 0; i < 3; i++ {
        //        soxPlay("/home/rpi3b/Desktop/VW_sounds/BRA_VW_SAVEIRO_INSTRUMENT_SOUNDS/VW_SVRO_DOT_BEEP.flac")
        //	time.Sleep(300 * time.Millisecond)
        //}
        //time.Sleep(450 * time.Millisecond)

        //for i:= 0; i < 3; i++ {
                soxPlay("/home/rpi3b/Desktop/VW_sounds/BRA_VW_SAVEIRO_INSTRUMENT_SOUNDS/VW_SVRO_MEDIA_NOTIFICATION.flac")
        //	time.Sleep(100 * time.Millisecond)
        	
        //}

	//algSay("Listening for GPS tracker")
	glob_date = "unavailable-"
	glob_location = "unavailable-"
	glob_counter = 0;
	go gpsTracker()	

	//algSay("Starting dash cam")

	

	for {
		loopCamera()
		time.Sleep(time.Millisecond * 200)
	}
}
