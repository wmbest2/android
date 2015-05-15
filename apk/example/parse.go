// Parse out
// package and activity(MAIN) from apk
package main

import (
	"archive/zip"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/nick-fedesna/android/apk"
)

func parseManifest(data []byte) {
	var manifest apk.Manifest
	err := apk.Unmarshal(data, &manifest)
	if err != nil {
		log.Fatal(err)
	}
	var mainAct *apk.AppActivity
	for _, act := range manifest.App.Activity {
		if mainAct != nil {
			break
		}
		for _, intent := range act.IntentFilter {
			if intent.Action.Name == "android.intent.action.MAIN" {
				mainAct = &act
				break
			}
		}
	}
	log.Println(manifest.Package)
	log.Println(mainAct.Name)
	fmt.Printf("adb shell am start -n %s/%s\n", manifest.Package, mainAct.Name)
	//out, _ := xml.MarshalIndent(manifest, "", "\t")
	//fmt.Printf("%s\n", string(out))
}

func ReadManifestFromApk(filename string) (data []byte, err error) {
	r, err := zip.OpenReader(filename)
	if err != nil {
		return
	}
	defer r.Close()
	for _, f := range r.File {
		if f.Name != "AndroidManifest.xml" {
			continue
		}
		rc, er := f.Open()
		if er != nil {
			return nil, er
		}
		data, err = ioutil.ReadAll(rc)
		rc.Close()
		return
	}
	return nil, fmt.Errorf("File not found: AndroidManifest.xml")
}

func main() {
	data, err := ReadManifestFromApk("ScreenTest.apk")
	if err != nil {
		log.Fatal(err)
	}
	parseManifest(data)
}
