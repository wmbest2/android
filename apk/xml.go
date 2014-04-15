package apk

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"math"
    "bytes"
    "encoding/xml"
)

const (
	CHUNK_AXML_FILE           = 0x00080003
	CHUNK_RESOURCEIDS         = 0x00080180
	CHUNK_STRINGS             = 0x001C0001
	CHUNK_XML_END_NAMESPACE   = 0x00100101
	CHUNK_XML_END_TAG         = 0x00100103
	CHUNK_XML_START_NAMESPACE = 0x00100100
	CHUNK_XML_START_TAG       = 0x00100102
	CHUNK_XML_TEXT            = 0x00100104
	UTF8_FLAG                 = 0x00000100
	SKIP_BLOCK                = 0xFFFFFFFF
)

type stringsMeta struct {
	Nstrings         uint32
	StyleOffsetCount uint32
	Flags            uint32
	StringDataOffset uint32
	Stylesoffset     uint32
    DataOffset		[]uint32
}

// decompressXML -- Parse the 'compressed' binary form of Android XML docs
// such as for AndroidManifest.xml in .apk files
func Unmarshal(b []byte, manifest interface{}) error {

    body := bytes.NewReader(b)

	var blocktype, size, indent, header uint32
    var stringsData stringsMeta

    // Check Header
	binary.Read(body, binary.LittleEndian, &header)
	if header != CHUNK_AXML_FILE {
		return errors.New("AXML file has wrong header")
	}

    // Check filesize
	binary.Read(body, binary.LittleEndian, &header)
	if int(header) != len(b) {
		return errors.New("AXML file has the wrong size")
	}

    var output string
	// Start offset at 8 bytes for header and size
	for offset := uint32(8); offset < header; {
		var lineNumber, skip, nsIdx, nameIdx, flag uint32
		binary.Read(body, binary.LittleEndian, &blocktype)
		binary.Read(body, binary.LittleEndian, &size)
        if  blocktype != CHUNK_RESOURCEIDS && blocktype != CHUNK_STRINGS {
			binary.Read(body, binary.LittleEndian, &lineNumber)
			binary.Read(body, binary.LittleEndian, &skip)
			if skip != SKIP_BLOCK {
				return errors.New("Error: Expected block 0xFFFFFFFF")
			}
			binary.Read(body, binary.LittleEndian, &nsIdx)
			binary.Read(body, binary.LittleEndian, &nameIdx)
			binary.Read(body, binary.LittleEndian, &flag)
        }
		switch blocktype {
		default:
			return fmt.Errorf("Unkown chunk type: %X", blocktype)
		case CHUNK_RESOURCEIDS:
		case CHUNK_STRINGS:
			/* +------------------------------------+
			 * | Nstrings         uint32            |
			 * | StyleOffsetCount uint32            |
			 * | Flags            uint32            |
			 * | StringDataOffset uint32            |
			 * | flag             uint32            |
			 * | Stylesoffset     uint32            |
			 * +------------------------------------+
			 * | +--------------------------------+ |
			 * | | DataOffset uint32              | |
			 * | +--------------------------------+ |
			 * |       Repeat Nstrings times        |
			 * +------------------------------------+
			 * |
			 * +------------------------------------+
			 */
			binary.Read(body, binary.LittleEndian, &stringsData.Nstrings)
			binary.Read(body, binary.LittleEndian, &stringsData.StyleOffsetCount)
			binary.Read(body, binary.LittleEndian, &stringsData.Flags)
			binary.Read(body, binary.LittleEndian, &stringsData.StringDataOffset)
			binary.Read(body, binary.LittleEndian, &stringsData.Stylesoffset)

            for i := uint32(0); i < stringsData.Nstrings; i++ {
				var offset uint32
				binary.Read(body, binary.LittleEndian, &offset)
                stringsData.DataOffset = append(stringsData.DataOffset, offset)
			}manifest
            stringsData.StringDataOffset = 0x24 + stringsData.Nstrings * 4
		case CHUNK_XML_END_NAMESPACE:
		case CHUNK_XML_END_TAG:
			indent--
			name := compXmlStringAt(body, stringsData, nameIdx)
			output = fmt.Sprintf("%s%s</%s>\n", output, computeIndent(indent), name)
		case CHUNK_XML_START_NAMESPACE:
		case CHUNK_XML_START_TAG:
			/* +----------------------------- w-------+
			 * | lineNumber     uint32              |
			 * | skip           uint32 = SKIP_BLOCK |
			 * | nsIdx          uint32              |
			 * | nameIdx        uint32              |
			 * | flag           uint32 = 0x00140014 |
			 * | attributeCount uint16              |
			 * +------------------------------------+
			 * | +--------------------------------+ |
			 * | | nsIdx       uint32             | |
			 * | | nameIdx     uint32             | |
			 * | | valueString uint32 // Skipped  | |
			 * | | aValueType  uint32             | |
			 * | | aValue      uint32             | |
			 * | +--------------------------------+ |
			 * |   Repeat attributeCount times      |
			 * +------------------------------------+
			 */

			var attributeCount, junk uint32
			// Check if flag is magick number
			// https://code.google.com/p/axml/source/browse/src/main/java/pxb/android/axml/AxmlReader.java?r=9bc9e64ef832736a93750998a9fa1d4406b858c3#102
			if flag != 0x00140014 {
				return fmt.Errorf("Expected flag 0x00140014, found %08X at %08X\n", flag, offset+4*6)
			}

            name := compXmlStringAt(body, stringsData, nameIdx)

			binary.Read(body, binary.LittleEndian, &attributeCount)
            binary.Read(body, binary.LittleEndian, &junk)

			var att string
			// Look for the Attributes
            for i := 0; i < int(attributeCount); i++ {
                var attrNameSi, attrNSSi, attrValueSi, flags, attrResId uint32 
                binary.Read(body, binary.LittleEndian, &attrNSSi)
                binary.Read(body, binary.LittleEndian, &attrNameSi)
                binary.Read(body, binary.LittleEndian, &attrValueSi)
                binary.Read(body, binary.LittleEndian, &flags)
                binary.Read(body, binary.LittleEndian, &attrResId)

				attrName := compXmlStringAt(body, stringsData, attrNameSi)

				var attrValue string
				if attrValueSi != 0xffffffff {
					attrValue = compXmlStringAt(body, stringsData, attrValueSi)
				} else {
					attrValue = fmt.Sprintf("%d", attrResId)
				}
				att = fmt.Sprintf("%s %s=\"%s\"", att, attrName, attrValue)
			}

			output = fmt.Sprintf("%s%s<%s%s>\n", output, computeIndent(indent), name, att)
			indent++
		case CHUNK_XML_TEXT:
		}
		offset += size
		body.Seek(int64(offset), 0)
	}
    return xml.Unmarshal([]byte(output), manifest)
}

func computeIndent(indent uint32) string {
	spaces := string("                                             ")
	m := int(math.Min(float64(indent*2), float64(len(spaces))))
	return spaces[:m]
}

// compXmlStringAt -- Return the string stored in StringTable format at
// offset strOff.  This offset points to the 16 bit string length, which
// is followed by that number of 16 bit (Unicode) chars.
func compXmlStringAt(arr io.ReaderAt, meta stringsMeta, strOff uint32) string {
    if strOff == 0xffffffff {
        return ""
    }
    length := make([]byte, 2)
    off := meta.StringDataOffset + meta.DataOffset[strOff]
	arr.ReadAt(length, int64(off))
	strLen := int(length[1] << 8 + length[0])

	chars := make([]byte, int64(strLen))
    ii := 0
    for i := 0; i < strLen; i++ {
        c := make([]byte, 1)
        arr.ReadAt(c, int64(int(off) + 2 + ii))

        if c[0] == 0 {
            i--
        } else {
            chars[i] = c[0]
        }
        ii++
	}

	return string(chars) 
} // end of compXmlStringAt
