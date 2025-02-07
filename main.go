package main

import (
	"flag"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"os"
	"path/filepath"
	"sort"
	"sync"
)

// Rectangle represents an image with an ID, width, height, and the image data itself.
type Rectangle struct {
	ID     int
	Width  int
	Height int
	Image  image.Image
}

// Shelf represents a horizontal shelf for packing rectangles in the texture atlas.
type Shelf struct {
	Y      int
	Height int
	Width  int
}

// main is the entry point of the program. It parses command-line flags,
// collects image files from a directory, loads and processes them,
// generates a texture atlas, saves it as 'atlas.png', and prints atlas information.
func main() {
	maxHeight, filedir := parseFlags()
	files, err := collectImageFiles(filedir)
	if err != nil {
		fmt.Println("Error collecting image files:", err)
		return
	}

	rectangles, err := loadImages(files)
	if err != nil {
		fmt.Println("Error loading images:", err)
		return
	}

	atlas, packedRectangles := generateAtlas(rectangles, maxHeight)

	if err := saveAtlas("atlas.png", atlas); err != nil {
		fmt.Println("Error saving atlas:", err)
		return
	}

	printAtlasInfo(atlas.Bounds().Max.X, atlas.Bounds().Max.Y, packedRectangles)
}

// parseFlags parses command-line flags to retrieve the maximum height
// of the texture atlas and the directory containing image files.
func parseFlags() (int, string) {
	maxHeight := flag.Int("maxheight", 1080, "Maximum height of the texture atlas")
	filedir := flag.String("filedir", "", "Directory containing image files")
	flag.Parse()

	if *filedir == "" {
		fmt.Println("Please provide a valid directory using -filedir flag.")
		os.Exit(1)
	}

	return *maxHeight, *filedir
}

// collectImageFiles retrieves a list of image files from the specified directory.
func collectImageFiles(filedir string) ([]string, error) {
	var files []string
	err := filepath.Walk(filedir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && isImageFile(path) {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

// isImageFile checks if the given filename has a supported image file extension.
func isImageFile(filename string) bool {
	switch filepath.Ext(filename) {
	case ".png", ".jpg", ".jpeg", ".gif", ".bmp":
		return true
	default:
		return false
	}
}

// loadImages loads image files concurrently, sorts them by height,
// and returns a slice of rectangles representing each loaded image.
func loadImages(files []string) ([]Rectangle, error) {
	rectangles := make([]Rectangle, len(files))
	var wg sync.WaitGroup
	errChan := make(chan error, len(files))

	for i, file := range files {
		wg.Add(1)
		go func(i int, file string) {
			defer wg.Done()
			img, err := loadImage(file)
			if err != nil {
				errChan <- fmt.Errorf("failed to load image %s: %w", file, err)
				return
			}
			rectangles[i] = Rectangle{
				ID:     i + 1,
				Image:  img,
				Width:  img.Bounds().Dx(),
				Height: img.Bounds().Dy(),
			}
		}(i, file)
	}

	wg.Wait()
	close(errChan)

	if err := <-errChan; err != nil {
		return nil, err
	}

	sort.Slice(rectangles, func(i, j int) bool {
		return rectangles[i].Height > rectangles[j].Height
	})

	return rectangles, nil
}

// loadImages loads image files concurrently, sorts them by height,
// and returns a slice of rectangles representing each loaded image.
func loadImage(file string) (image.Image, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	return img, nil
}

// generateAtlas packs the provided rectangles into a texture atlas image
// using a shelf packing algorithm and returns the texture atlas image
// along with the mapping of rectangle IDs to their positions in the atlas.
func generateAtlas(rectangles []Rectangle, maxHeight int) (*image.RGBA, map[int]image.Rectangle) {
	packedRectangles := make(map[int]image.Rectangle)
	shelves := []Shelf{{Y: 0, Height: 0, Width: 0}}
	maxWidth := 0

	for _, rect := range rectangles {
		packed := false
		for i, shelf := range shelves {
			if rect.Height <= shelf.Height && shelf.Width+rect.Width <= maxHeight {
				packedRectangles[rect.ID] = image.Rect(shelf.Width, shelf.Y, shelf.Width+rect.Width, shelf.Y+rect.Height)
				shelves[i].Width += rect.Width
				if shelves[i].Width > maxWidth {
					maxWidth = shelves[i].Width
				}
				packed = true
				break
			}
		}

		if !packed {
			newShelf := Shelf{Y: shelves[len(shelves)-1].Y + shelves[len(shelves)-1].Height, Height: rect.Height, Width: rect.Width}
			shelves = append(shelves, newShelf)
			packedRectangles[rect.ID] = image.Rect(0, newShelf.Y, rect.Width, newShelf.Y+rect.Height)
			if rect.Width > maxWidth {
				maxWidth = rect.Width
			}
		}
	}

	totalHeight := shelves[len(shelves)-1].Y + shelves[len(shelves)-1].Height
	atlas := image.NewRGBA(image.Rect(0, 0, maxWidth, totalHeight))

	for _, rect := range rectangles {
		draw.Draw(atlas, packedRectangles[rect.ID], rect.Image, image.Point{}, draw.Src)
	}

	return atlas, packedRectangles
}

// saveAtlas saves the texture atlas image as a PNG file with the specified filename.
func saveAtlas(filename string, atlas *image.RGBA) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	encoder := png.Encoder{CompressionLevel: png.BestCompression}
	return encoder.Encode(f, atlas)
}

// printAtlasInfo prints information about the generated texture atlas,
// including its dimensions and the positions of packed rectangles.
func printAtlasInfo(width, height int, packedRectangles map[int]image.Rectangle) {
	fmt.Printf("Atlas size: %d x %d\n", width, height)
	fmt.Println("Packed rectangles:")
	for id, rect := range packedRectangles {
		fmt.Printf("ID: %d, Rect: %v\n", id, rect)
	}
	fmt.Println("Atlas saved as atlas.png successfully.")
}
