package main

import (
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"path/filepath"
)

// Point structure (matching generated code)
type Point struct {
	X float64
	Y float64
}

// Rectangle structure (matching generated code)
type Rectangle struct {
	TopLeft Point
	Width   float64
	Height  float64
}

// Message header constants (matching generated code)
const (
	MessageMagic   = "SDP" // 3 bytes: 'S', 'D', 'P'
	MessageVersion = '2'   // ASCII '2' for v0.2.0
)

// Type IDs (matching schema order)
const (
	TypeIDPoint     uint16 = 1
	TypeIDRectangle uint16 = 2
)

// Encode Point to buffer
func encodePoint(p *Point, buf []byte) {
	binary.LittleEndian.PutUint64(buf[0:8], math.Float64bits(p.X))
	binary.LittleEndian.PutUint64(buf[8:16], math.Float64bits(p.Y))
}

// Encode Rectangle to buffer
func encodeRectangle(r *Rectangle, buf []byte) {
	encodePoint(&r.TopLeft, buf[0:16])
	binary.LittleEndian.PutUint64(buf[16:24], math.Float64bits(r.Width))
	binary.LittleEndian.PutUint64(buf[24:32], math.Float64bits(r.Height))
}

// Encode Point message (with header)
func encodePointMessage(p *Point) []byte {
	const payloadSize = 16
	const headerSize = 10 // SDP(3) + version(1) + type_id(2) + length(4)
	const totalSize = headerSize + payloadSize
	
	buf := make([]byte, totalSize)
	
	// Header: [SDP:3][version:1][type_id:2][length:4]
	copy(buf[0:3], MessageMagic)
	buf[3] = MessageVersion
	binary.LittleEndian.PutUint16(buf[4:6], TypeIDPoint)
	binary.LittleEndian.PutUint32(buf[6:10], payloadSize)
	
	// Payload starts at byte 10
	encodePoint(p, buf[10:])
	
	return buf
}

// Encode Rectangle message (with header)
func encodeRectangleMessage(r *Rectangle) []byte {
	const payloadSize = 32
	const headerSize = 10 // SDP(3) + version(1) + type_id(2) + length(4)
	const totalSize = headerSize + payloadSize
	
	buf := make([]byte, totalSize)
	
	// Header: [SDP:3][version:1][type_id:2][length:4]
	copy(buf[0:3], MessageMagic)
	buf[3] = MessageVersion
	binary.LittleEndian.PutUint16(buf[4:6], TypeIDRectangle)
	binary.LittleEndian.PutUint32(buf[6:10], payloadSize)
	
	// Payload starts at byte 10
	encodeRectangle(r, buf[10:])
	
	return buf
}

func main() {
	// Create binaries directory if it doesn't exist
	binariesDir := filepath.Join("..", "binaries")
	if err := os.MkdirAll(binariesDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating binaries directory: %v\n", err)
		os.Exit(1)
	}
	
	// Encode Point message
	point := Point{
		X: 3.14,
		Y: 2.71,
	}
	
	pointData := encodePointMessage(&point)
	
	pointPath := filepath.Join(binariesDir, "message_point.sdpb")
	if err := os.WriteFile(pointPath, pointData, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing Point file: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Created %s (%d bytes)\n", pointPath, len(pointData))
	
	// Print hex for verification
	fmt.Printf("  Hex: ")
	for _, b := range pointData {
		fmt.Printf("%02x", b)
	}
	fmt.Println()
	
	// Encode Rectangle message
	rect := Rectangle{
		TopLeft: Point{X: 10.0, Y: 20.0},
		Width:   100.0,
		Height:  50.0,
	}
	
	rectData := encodeRectangleMessage(&rect)
	
	rectPath := filepath.Join(binariesDir, "message_rectangle.sdpb")
	if err := os.WriteFile(rectPath, rectData, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing Rectangle file: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Created %s (%d bytes)\n", rectPath, len(rectData))
	
	// Print hex for verification
	fmt.Printf("  Hex: ")
	for _, b := range rectData {
		fmt.Printf("%02x", b)
	}
	fmt.Println()
	
	fmt.Println("\nReference .sdpb files generated successfully!")
}
