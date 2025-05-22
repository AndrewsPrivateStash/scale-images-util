<pre>
A simple image resizing util, that recursively finds image files of several types {jpeg, png, tiff, webp, bmp}
and resizes the image towards a target. Upscaling up to 1.5 increase ratio; and any degree of decrease needed.
A very small window around the target will result in a no-op and the image will be copied instead without conversion to jpeg.

all encodings are jpeg, with default quality 90 (1-100). You can change this with the -q flag.
the default target is 1440 x 2560 = 3,686,400 pixels

the dimension ratio factor is calculated as: r = sqrt(t / T)
    where t is the target pixel count, and T is the image pixel count
    both dimensions are scaled by this ratio, subject to increase constraints

When the -f flag is not set; the src directory structure will be re-produced in the out dir defined by -o
    the out_dir will only contain roots such that an image exists in that tree

The CatmullRom kernel is used for interpolation, which is very slow, but produces nice resizing results.
https://pkg.go.dev/golang.org/x/image/draw#pkg-variables


</pre>

```
Usage of resizeImgs:
  -f	extract images to a single dir ignoring src structure
  -o string
    	out path (default "out_imgs")
  -q int
    	jpeg quality 1-100 (default 90)
  -r	recursive process
  -t int
    	target pixels (default 3686400)
```
<pre>
Examples:
$ resizeImgs -h                  help
$ resizeImgs -r -f [root_dir]    resize all images contained in the root_dir tree and store in flat dir
$ resizeImgs -q 50 -t 2073600 [file_path]    resize single file with jpeg quality 50 to target 1080x1920
$ resizeImgs -r -f -o my_pics /home/[user]   resize all pics from your home tree storing in dir my_pics
</pre>