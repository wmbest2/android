package apk

import (
    "fmt"
    "math"
    "encoding/xml"
)
// decompressXML -- Parse the 'compressed' binary form of Android XML docs 
// such as for AndroidManifest.xml in .apk files
func Unmarshal(body []byte, manifest *Manifest) error {
    endDocTag := 0x00100101
    startTag :=  0x00100102
    endTag := 0x00100103

// Compressed XML file/bytes starts with 24x bytes of data,
    // 9 32 bit words in little endian order (LSB first):
    //   0th word is 03 00 08 00
    //   3rd word SEEMS TO BE:  Offset at then of StringTable
    //   4th word is: Number of strings in string table
    // WARNING: Sometime I indiscriminently display or refer to word in 
    //   little endian storage format, or in integer format (ie MSB first).
    numbStrings := lew(body, 4*4)

    // StringIndexTable starts at offset 24x, an array of 32 bit LE offsets
    // of the length/string data in the StringTable.
    sitOff := 0x24  // Offset of start of StringIndexTable

    // StringTable, each string is represented with a 16 bit little endian 
    // character count, followed by that number of 16 bit (LE) (Unicode) chars.
    stOff := sitOff + numbStrings*4  // StringTable follows StrIndexTable

    // XMLTags, The XML tag tree starts after some unknown content after the
    // StringTable.  There is some unknown data after the StringTable, scan
    // forward from this point to the flag for the start of an XML start tag.
    bodyTagOff := lew(body, 3*4)  // Start from the offset in the 3rd word.
    // Scan forward until we find the bytes: 0x02011000(x00100102 in normal int)
    for ii := bodyTagOff; ii < len(body)-4; ii += 4 {
       if (lew(body, ii) == startTag) { 
           bodyTagOff = ii
           break
       }
    } // end of hack, scanning for start of first start tag

    // XML tags and attributes:
    // Every XML start and end tag consists of 6 32 bit words:
    //   0th word: 02011000 for startTag and 03011000 for endTag 
    //   1st word: a flag?, like 38000000
    //   2nd word: Line of where this tag appeared in the original source file
    //   3rd word: FFFFFFFF ??
    //   4th word: StringIndex of NameSpace name, or FFFFFFFF for default NS
    //   5th word: StringIndex of Element Name
    //   (Note: 01011000 in 0th word means end of XML document, endDocTag)

    // Start tags (not end tags) contain 3 more words:
    //   6th word: 14001400 meaning?? 
    //   7th word: Number of Attributes that follow this tag(follow word 8th)
    //   8th word: 00000000 meaning??

    // Attributes consist of 5 words: 
    //   0th word: StringIndex of Attribute Name's Namespace, or FFFFFFFF
    //   1st word: StringIndex of Attribute Name
    //   2nd word: StringIndex of Attribute Value, or FFFFFFF if ResourceId used
    //   3rd word: Flags?
    //   4th word: str ind of attr value again, or ResourceId of value


    // Step through the XML tree element tags and attributes
    var output string
    off := bodyTagOff
    indent := 0
    for off < len(body) {
        tag0 := lew(body, off)
        nameSi := lew(body, off+5*4)

        if (tag0 == startTag) { // XML START TAG
            numbAttrs := lew(body, off+7*4)  // Number of Attributes to follow
            off += 9*4;  // Skip over 6+3 words of startTag data
            name := compXmlString(body, sitOff, stOff, nameSi)

            var att string
            // Look for the Attributes
            for ii := 0; ii < numbAttrs; ii++ {
                attrNameSi := lew(body, off+1*4)  // AttrName String Index
                attrValueSi := lew(body, off+2*4) // AttrValue Str Ind, or FFFFFFFF
                attrResId := lew(body, off+4*4)  // AttrValue ResourceId or dup AttrValue StrInd
                off += 5*4;  // Skip over the 5 words of an attribute

                attrName := compXmlString(body, sitOff, stOff, attrNameSi)

                var attrValue string
                if attrValueSi != 0xffffffff {
                    attrValue = compXmlString(body, sitOff, stOff, attrValueSi)
                } else {
                    attrValue = fmt.Sprintf("%d", attrResId)
                }
                att = fmt.Sprintf("%s %s=\"%s\"", att, attrName, attrValue)
            }

            output = fmt.Sprintf("%s%s<%s%s>\n", output, computeIndent(indent), name, att)
            indent++;
        } else if (tag0 == endTag) { // XML END TAG
            indent--
            off += 6*4  // Skip over 6 words of endTag data
            name := compXmlString(body, sitOff, stOff, nameSi)
            output = fmt.Sprintf("%s%s</%s>\n", output, computeIndent(indent), name)

        } else if (tag0 == endDocTag) {  // END OF XML DOC TAG
            break;
        } else {
            output = fmt.Sprintf("%s  Unrecognized tag code '%d' at offset %d\n", output, tag0, off)
            break;
        }
    } // end of while loop scanning tags and attributes of XML tree
    //fmt.Printf("%s    end at offset %d\n", output, off)
    return xml.Unmarshal([]byte(output), manifest)
} // end of decompressXML


func compXmlString(xml []byte, sitOff int, stOff int, strInd int) string{
  if strInd < 0  {
      return "";
  }
  strOff := stOff + lew(xml, sitOff+strInd*4)
  return compXmlStringAt(xml, strOff)
}

func computeIndent(indent int) string {
    spaces := string("                                             ")
    m := int(math.Min(float64(indent*2), float64(len(spaces))))
    return spaces[:m]
}


// compXmlStringAt -- Return the string stored in StringTable format at
// offset strOff.  This offset points to the 16 bit string length, which 
// is followed by that number of 16 bit (Unicode) chars.
func compXmlStringAt(arr []byte, strOff int) string {
  strLen := int(arr[strOff+1])<<8&0xff00 | int(arr[strOff])&0xff;

  chars := make([]byte, strLen)
  for i,_ := range chars {
    chars[i] = arr[strOff+2+i*2];
  }
  return string(chars)  // Hack, just use 8 byte chars
} // end of compXmlStringAt


// lew -- Return value of a Little Endian 32 bit word from the byte array
//   at offset off.
func lew(arr []byte, off int) int {
  return int(arr[off+3])<<24&0xff000000 | int(arr[off+2])<<16&0xff0000 | int(arr[off+1])<<8&0xff00 | int(arr[off]) & 0xFF
} // end of lew
