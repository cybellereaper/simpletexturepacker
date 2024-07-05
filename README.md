# Simple Texture Packer

This command-line tool generates a texture atlas from a directory of image files. A texture atlas is a large image that contains several smaller images packed together, which is commonly used in graphics applications to optimize rendering.

## Installation

To install the texture atlas generator, clone the repository:

`git clone <repository-url>`
`cd <repository-directory>`

## Usage

### Command-line Flags

- `-maxheight`: Maximum height of the texture atlas (default: 1080).
- `-filedir`: Directory containing the image files (required).

### Example

Generate a texture atlas from images in the `images` directory with a maximum height of 1024 pixels:

`./texture-atlas-generator -maxheight 1024 -filedir ./images`

## How It Works

1. **Collect Image Files**: The tool recursively scans the specified directory for supported image files (PNG, JPG, JPEG, GIF, BMP).

2. **Load Images**: Images are loaded concurrently using goroutines and sorted by height in descending order to optimize packing.

3. **Generate Texture Atlas**: The tool packs the sorted images into a single texture atlas image, arranging them to minimize wasted space.

4. **Save Atlas**: Finally, the texture atlas is saved as `atlas.png` in the current directory, and information about the packed rectangles is printed.

## Dependencies

- Go standard library packages (`image`, `image/draw`, `image/png`, `os`, `flag`, `filepath`, `fmt`, `sync`).

## Contributing

Contributions are welcome! If you have suggestions or improvements, please open an issue or pull request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.
