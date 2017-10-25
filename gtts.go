package main

import (
	"regexp"
	"github.com/zaf/agi"
	"github.com/nubunto/tts"
	"io/ioutil"
	"log"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"crypto/md5"
)

type codec struct{
	codec string
	rate string
}

const debug = false
const workDir = "/tmp/"
var codecs = map[string]codec{
	"silk12"	: codec{ "sln12","12000"},
	"sln12"		: codec{ "sln12","12000"},
	"speex16"	: codec{ "sln16", "16000"},
	"slin16"	: codec{ "sln16", "16000"},
	"silk16"	: codec{ "sln16", "16000"},
	"g722"		: codec{ "sln16", "16000"},
	"siren7"	: codec{ "sln16", "16000"},
	"speex32"	: codec{ "sln32", "32000"},
	"slin32"	: codec{ "sln32", "32000"},
	"celt32"	: codec{ "sln32", "32000"},
	"siren14"	: codec{ "sln32", "32000"},
	"celt44"	: codec{ "sln44", "44100"},
	"slin44"	: codec{ "sln44", "44100"},
	"celt48"	: codec{ "sln44", "44100"},
	"slin48"	: codec{ "sln44", "44100"},
	"other" 	: codec{ "sln", "8000"},
}

func main() {
	
	text := ""
	if len(os.Args) < 1{
		log.Fatalf("Nothing to translate to speech here" )
	}else{
		text = os.Args[1]
	}

	// Create a new AGI session and Parse the AGI environment.
	myAgi := agi.New()
	err := myAgi.Init(nil)
	
	if err != nil {
		log.Fatalf("Error Parsing AGI environment: %v\n", err)
	}

	// Handle Hangup from the asterisk server
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGHUP)
	go handleHangup(sigChan)
	
	//check/detect format
	myAgi.Verbose( myAgi.Env["channel"] )
	foos, err := myAgi.GetFullVariable( "${CHANNEL(audionativeformat)}",  myAgi.Env["channel"] )
	myAgi.Verbose( foos )
	myAgi.Verbose( foos.Dat )
	audioformat := "other"
	re := regexp.MustCompile("((silk|sln)12)|((speex|slin|silk)16|g722|siren7)|((speex|slin|celt)32|siren14)|((celt|slin)44)|((celt|slin)48)")
	if re.FindString( foos.Dat ) != "" {
		audioformat = re.FindString( foos.Dat )
	}

	//Playback recieved text Message
	myAgi.Verbose( playback(text, audioformat, myAgi ))
}
//play text to asterisk
func playback(text string, format string, myAgi *agi.Session) string{

	name := hashString( text )
	astName := name

	if alreadyCached( name, format ) == false {
		mp3Name := getText2Speach( text, name )
		wavName := convert2Wav( mp3Name )
		astName = convert2Aster( wavName, format )
	}else{
		astName = name
	}

	rep, err := myAgi.StreamFile( workDir+astName, "0123456789")

	if err != nil {
		log.Fatalf("AGI reply error: %v\n", err)
	}
	if rep.Res == -1 {
		log.Printf("Error during playback\n")
	}
	return astName
}
//hashout a string we'll use this for a filename
func hashString( text string ) string{
	return fmt.Sprintf("%x", md5.Sum([]byte(text)))
}
//check if we already have an audio file cached with the string we're translating
func alreadyCached( filename string, audioformat string ) bool{
	if _, err := os.Stat(workDir+filename+"."+codecs[ audioformat ].codec); os.IsNotExist(err) {
		return false
	}
	return true
}
//do our text 2 speach dealie... retrieved as an mp3
func getText2Speach( text string, filename string ) string{

	//get our message from google translate
	message := text
	s, err := tts.Speak(tts.Config{
		Speak:    message,
		Language: "en",
	})
	if err != nil {
		log.Fatal(err)
	}

	//write out our audio file
	err = ioutil.WriteFile(workDir+filename+".mp3", s.Bytes(), os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}
	return filename
}
//use mpg123 to convert our recieved mp3 to a wav
func convert2Wav( filename string ) string{

	//mpg123 -q -w <wav_output_file_name> <file_to_convert>
	result, err := exec.Command("mpg123", "-q", "-w", workDir+filename+".wav", workDir+filename+".mp3").Output()

	if err != nil {
		log.Fatalf("error converting mp3 to wav: %v\n", err)
	}else if debug ==  true{
		fmt.Println(result)
	}
	return filename
}
//use sox to convert to an asterisk friendly sound format
func convert2Aster( filename string, audioformat string ) string{

	result, err := exec.Command( "sox", workDir+filename+".wav", "-q", "-r", codecs[ audioformat ].rate, "-t", "raw", workDir+filename+"."+codecs[ audioformat ].codec ).Output()

	if err != nil {
		log.Fatalf("error transcoding with sox: %v\n", err)
	}else{
		os.Remove(workDir+filename+".mp3")
		os.Remove(workDir+filename+".wav")
	}
	if debug == true{
		fmt.Println(result)
	}
	return  filename
}
//hang up the line if we hang up
func handleHangup(sch <-chan os.Signal) {
	signal := <-sch
	log.Printf("Received %v, exiting...\n", signal)
	os.Exit(1)
}
