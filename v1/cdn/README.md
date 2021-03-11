
# CDN Documentation
Sign URLs with key and salt, so an attacker won’t be able to cause a denial-of-service attack by requesting multiple different image resizes.

## Features
- Generate URL to image proxy
- Generate s3 URL

## Quickstart

### Using s3 path
```go
import (
"fmt"
"github.com/forkyid/go-utils/v1/cdn"
)

  

func ExampleClient() {
	c, _ := cdn.New("http://example.com", "key", "saltkey")
	
	url := c.GetUrl(&cdn.Image{
						Url: c.GetS3Url(&cdn.S3{
							BucketName: "bucket.example.com",
							Path: "users/example/post/photos/5ee87fad6b12c90001cf41ab.jpg",}),
						Width: 800,
						})
	fmt.Println(url)
	// Output: http://example.com/Gc3qJZz3REIrYcZxqY1oXiTNSAplq8fhLxgrIqybLlA/fill/800/0/no/1/czM6Ly9pbWcuZm9ya3kuaWQvdXNlcnMvYWd1c2R3aS9wb3N0L3Bob3Rvcy81ZWU4N2ZhZDZiMTJjOTAwMDFjZjQxYWIuanBn.webp
}
```

##
### Using URL

```go
import (
"fmt"
"github.com/forkyid/go-utils/v1/cdn"
)

  

func ExampleClient() {
	c, _ := cdn.New("http://example.com", "key", "saltkey")
	
	url := c.GetUrl(&cdn.Image{
						Url: "https://example.com/5ee87fad6b12c90001cf41ab.jpg",
						Width: 800,
						})
	fmt.Println(url)
	// Output: http://example.com/Gc3qJZz3REIrYcZxqY1oXiTNSAplq8fhLxgrIqybLlA/fill/800/0/no/1/czM6Ly9pbWcuZm9ya3kuaWQvdXNlcnMvYWd1c2R3aS9wb3N0L3Bob3Rvcy81ZWU4N2ZhZDZiMTJjOTAwMDFjZjQxYWIuanBn.webp
}
```

##
### Image Configuration
```go
type  Image  struct {
	Url string
	Resize string
	Width int
	Height int
	X int
	Y int
	Gravity string
	Enlarge int
	Extension string
}
```

#### Information:
- Resizing types
Supports the following resizing types:
	-  `fit`: resizes the image while keeping aspect ratio to fit given size;
	- `fill`: resizes the image while keeping aspect ratio to fill given size and cropping projecting parts;
	-  `auto`: if both source and resulting dimensions have the same orientation (portrait or landscape), imgproxy will use  `fill`. Otherwise, it will use  `fit`.

- Width and height :
Width and height parameters define the size of the resulting image in pixels. Depending on the resizing type applied, the dimensions may differ from the requested ones.

- Gravity
When imgproxy needs to cut some parts of the image, it is guided by the gravity. The following values are supported:
	-   `no`: north (top edge);
	-   `so`: south (bottom edge);
	-   `ea`: east (right edge);
	-   `we`: west (left edge);
	-   `noea`: north-east (top-right corner);
	-   `nowe`: north-west (top-left corner);
	-   `soea`: south-east (bottom-right corner);
	-   `sowe`: south-west (bottom-left corner);
	-   `ce`: center;
	-   `sm`: smart.  `libvips`  detects the most “interesting” section of the image and considers it as the center of the resulting image;
	-   `fp:%x:%y`  - focus point.  `x`  and  `y`  are floating point numbers between 0 and 1 that describe the coordinates of the center of the resulting image. Treat 0 and 1 as right/left for  `x`  and top/bottom for  `y`.

- Enlarge
When set to  `1`,  `t`  or  `true`, imgproxy will enlarge the image if it is smaller than the given size.

- URL
There are two ways to specify source url:
	- S3
		```go
		c.GetUrl(&cdn.Image{
			Url: c.GetS3Url(&cdn.S3{
					BucketName: "bucket.example.com",
					Path: "users/example/post/photos/5ee87fad6b12c90001cf41ab.jpg",
				}),
			Width:  800,
		})
		```
	- Image URL
		 ```go
		c.GetUrl(&cdn.Image{
			Url: "https://example.com/5ee87fad6b12c90001cf41ab.jpg",
			Width:  800,
		})
		```

- Extension
Extension specifies the format of the resulting image. At the moment, imgproxy supports only  `jpg`,  `png`,  `webp`,  `gif`,  `ico`, and  `tiff`, them being the most popular and useful image formats.