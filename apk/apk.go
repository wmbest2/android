package apk

/* 
    <manifest xmlns:android="http://schemas.android.com/apk/res/android"
          package="string"
          android:sharedUserId="string"
          android:sharedUserLabel="string resource" 
          android:versionCode="integer"
          android:versionName="string"
          android:installLocation=["auto" | "internalOnly" | "preferExternal"] >
    <application> 
    <compatible-screens> 
    <instrumentation> 
    <permission> 
    <permission-group> 
    <permission-tree> 
    <supports-gl-texture><supports-screens> 
    <uses-configuration> 
    <uses-feature> 
    <uses-permission> 
    <uses-sdk>
*/


/*
<instrumentation android:functionalTest=["true" | "false"]
     android:handleProfiling=["true" | "false"]
     android:icon="drawable resource"
     android:label="string resource"
     android:name="string"
     android:targetPackage="string" />
*/
type Instrumentation struct {
    Name string `xml:"name,attr"`
    Target string `xml:"targetPackage,attr"`
}

type Application struct {

}

type Manifest struct {
    Package string `xml:"package,attr"`
    VersionCode int `xml:"versionCode,attr"`
    VersionName string `xml:"versionName,attr"`
    App Application `xml:"application"`
    Instrument Instrumentation `xml:"instrumentation"`
}
