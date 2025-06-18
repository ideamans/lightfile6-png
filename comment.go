package png

import (
	"bytes"
	"encoding/json"
	"fmt"

	pngstructure "github.com/dsoprea/go-png-image-structure/v2"
)

type LightFileComment struct {
	By       string `json:"by"`       // Optimization tool identifier
	Before   int64  `json:"before"`   // Original file size in bytes
	After    int64  `json:"after"`    // Optimized file size in bytes
	PNGQuant bool   `json:"pngquant"` // Indicates if PNGQuant was used
}

// ReadComment reads a LightFileComment from PNG tEXt chunks.
// Returns:
//   - *LightFileComment: Parsed comment or nil if JSON parsing fails
//   - string: Raw text string from LightFile tEXt chunk, or "" if not found
//   - error: DataError if PNG structure is invalid, system error otherwise
func ReadComment(data []byte) (*LightFileComment, string, error) {
	pmp := pngstructure.NewPngMediaParser()

	mediaContext, err := pmp.ParseBytes(data)
	if err != nil {
		return nil, "", NewDataErrorf("failed to parse PNG structure: %v", err)
	}

	cs := mediaContext.(*pngstructure.ChunkSlice)
	chunks := cs.Chunks()
	
	for _, chunk := range chunks {
		if chunk.Type == "tEXt" {
			textData := chunk.Data
			
			// tEXt format: keyword\0text
			nullIndex := bytes.IndexByte(textData, 0)
			if nullIndex == -1 {
				continue
			}
			
			keyword := string(textData[:nullIndex])
			text := string(textData[nullIndex+1:])
			
			// Look for LightFile comment
			if keyword == "LightFile" {
				var comment LightFileComment
				err := json.Unmarshal([]byte(text), &comment)
				if err != nil {
					// Return raw text even if JSON parsing fails
					return nil, text, nil
				}
				return &comment, text, nil
			}
		}
	}
	
	return nil, "", nil
}

// BuildComment builds a JSON comment string from a LightFileComment.
// Returns:
//   - string: JSON representation of the comment
//   - int: Size of the JSON string
//   - error: System error if JSON marshaling fails
func BuildComment(comment *LightFileComment) (string, int, error) {
	jsonData, err := json.Marshal(comment)
	if err != nil {
		return "", 0, fmt.Errorf("failed to marshal comment to JSON: %w", err)
	}
	
	jsonStr := string(jsonData)
	return jsonStr, len(jsonStr), nil
}

// WriteComment writes a comment as JSON into PNG data using tEXt chunk.
// Returns:
//   - []byte: New PNG data with comment embedded
//   - error: DataError if PNG structure is invalid, system error otherwise
func WriteComment(data []byte, comment string) ([]byte, error) {
	pmp := pngstructure.NewPngMediaParser()

	mediaContext, err := pmp.ParseBytes(data)
	if err != nil {
		return nil, NewDataErrorf("failed to parse PNG structure: %v", err)
	}

	// Create tEXt chunk data
	keyword := "LightFile"
	textData := make([]byte, len(keyword)+1+len(comment))
	copy(textData, keyword)
	textData[len(keyword)] = 0 // null separator
	copy(textData[len(keyword)+1:], comment)

	// Find where to insert the tEXt chunk (before IEND)
	cs := mediaContext.(*pngstructure.ChunkSlice)
	chunks := cs.Chunks()
	newChunks := make([]*pngstructure.Chunk, 0, len(chunks)+1)
	
	inserted := false
	for _, chunk := range chunks {
		if chunk.Type == "IEND" && !inserted {
			// Insert our tEXt chunk before IEND
			textChunk := &pngstructure.Chunk{
				Type: "tEXt",
				Data: textData,
			}
			newChunks = append(newChunks, textChunk)
			inserted = true
		}
		newChunks = append(newChunks, chunk)
	}
	
	if !inserted {
		return nil, NewDataError("PNG file missing IEND chunk")
	}

	// Rebuild PNG with new chunks
	var buf bytes.Buffer
	
	// Write PNG signature
	buf.Write([]byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A})
	
	for _, chunk := range newChunks {
		err := writeChunk(&buf, chunk)
		if err != nil {
			return nil, fmt.Errorf("failed to write chunk: %w", err)
		}
	}
	
	return buf.Bytes(), nil
}

// writeChunk writes a PNG chunk to the buffer
func writeChunk(buf *bytes.Buffer, chunk *pngstructure.Chunk) error {
	// Write length (4 bytes, big-endian)
	length := uint32(len(chunk.Data))
	buf.WriteByte(byte(length >> 24))
	buf.WriteByte(byte(length >> 16))
	buf.WriteByte(byte(length >> 8))
	buf.WriteByte(byte(length))
	
	// Write type (4 bytes)
	buf.WriteString(chunk.Type)
	
	// Write data
	buf.Write(chunk.Data)
	
	// Calculate and write CRC (4 bytes)
	crcData := make([]byte, 4+len(chunk.Data))
	copy(crcData, chunk.Type)
	copy(crcData[4:], chunk.Data)
	crc := crc32PNG(crcData)
	
	buf.WriteByte(byte(crc >> 24))
	buf.WriteByte(byte(crc >> 16))
	buf.WriteByte(byte(crc >> 8))
	buf.WriteByte(byte(crc))
	
	return nil
}

// crc32PNG calculates CRC32 for PNG chunks
func crc32PNG(data []byte) uint32 {
	var crcTable [256]uint32
	
	// Initialize CRC table
	for i := 0; i < 256; i++ {
		c := uint32(i)
		for j := 0; j < 8; j++ {
			if c&1 == 1 {
				c = 0xEDB88320 ^ (c >> 1)
			} else {
				c = c >> 1
			}
		}
		crcTable[i] = c
	}
	
	crc := uint32(0xFFFFFFFF)
	for _, b := range data {
		crc = crcTable[(crc^uint32(b))&0xFF] ^ (crc >> 8)
	}
	return crc ^ 0xFFFFFFFF
}
