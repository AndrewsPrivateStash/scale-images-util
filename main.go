/*
	Andrew Pfaendler
	2025-05-18

	ToDo

*/

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"slices"
	"sync"
)

type dims struct {
	x int
	y int
}

type Args struct {
	inPath  string
	outPath string
	rec     bool
	target  int
	qual    int
	flat    bool
}

type Tracking = map[string]int

var TRK = Tracking{
	"scaled_cnt": 0,
	"copied_cnt": 0,
	"skipped":    0,
	"errored":    0,
}
var LOCK = sync.RWMutex{}
var wg = sync.WaitGroup{}

var TYPES = []string{"jpeg", "jpg", "bmp", "tiff", "webp", "png"}

func main() {

	var (
		outPathF = flag.String("o", "out_imgs", "out path")
		targetF  = flag.Int("t", 3686400, "target pixels") // 2560 x 1440
		qualF    = flag.Int("q", 90, "jpeg quality 1-100")
		recurseF = flag.Bool("r", false, "recursive process")
		flatOutF = flag.Bool("f", false, "extract images to a single dir ignoring src structure")
	)
	flag.Parse()

	if flag.NArg() < 1 {
		log.Fatal("expected path to process: $ resize_imgs my_imgs")
	}

	args := Args{
		inPath:  flag.Args()[0],
		outPath: *outPathF,
		rec:     *recurseF,
		target:  *targetF,
		qual:    *qualF,
		flat:    *flatOutF,
	}

	//	TEST

	//

	// remove existing dir if exists
	if err := remove_out_dir(*outPathF); err != nil {
		log.Fatal(err)
	}

	// make output folder
	if err := os.MkdirAll(*outPathF, 0755); err != nil {
		log.Fatal(err)
	}

	wg.Add(1)
	process_files(args.inPath, args.outPath, args, TRK)
	if args.rec {
		wg.Wait()
	}

	print_track_data(TRK)

}

func print_track_data(trk Tracking) {
	for k, v := range trk {
		if v > 0 {
			fmt.Printf("%s: %d file(s)\n", k, v)
		}
	}
}

func remove_out_dir(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil
	} else if err == nil {
		os.RemoveAll(path)
		return nil
	} else {
		return err
	}
}

func isFile(path string) bool {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false // Handle errors like file not found
	}
	return !fileInfo.IsDir()
}

// does the file have an extension we can handle
func canProc(path string) bool {
	ext := filepath.Ext(path)
	if ext != "" {
		return slices.Contains(TYPES, ext[1:])
	}
	return false
}

// does the root dir contain an img file anywhere in its tree
func treeHasImg(rootDir string, state bool) bool {
	if state {
		return true
	}

	files, err := os.ReadDir(rootDir)
	if err != nil || len(files) == 0 {
		return false
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if canProc(file.Name()) {
			return true
		}
	}

	for _, file := range files {
		if file.IsDir() {
			fnd := treeHasImg(filepath.Join(rootDir, file.Name()), false)
			if fnd {
				return true
			}
		}

	}

	return false

}

func load_img(img_path string, dest_dir string) *ImgObj {
	if !isFile(img_path) {
		return nil
	}

	fn := filepath.Base(img_path)
	if !canProc(fn) {
		return nil
	}

	return newImgObj(img_path, filepath.Join(dest_dir, fn))
}

func process_files(path string, out_dir string, a Args, trk Tracking) {

	// single file case
	if isFile(path) {
		log.Printf("processing single file: %s\t--->\t%s\n", path, a.outPath)
		img := load_img(path, a.outPath)
		if img == nil {
			log.Fatalf("did not load file\n%s\n", path)
		}
		if err := img.proc_img(a, trk); err != nil {
			log.Fatalf("error processing file: %s\n%s\n", img.src_path, err)
		}
		return
	}

	// directory case
	files, err := os.ReadDir(path)
	if err != nil {
		log.Fatalf("error processing directory files\n%s\n", err)
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if !canProc(file.Name()) {
			LOCK.Lock()
			trk["skipped"]++
			LOCK.Unlock()
			continue
		}

		img_path := filepath.Join(path, file.Name())
		img := load_img(img_path, out_dir)
		if img == nil {
			log.Printf("err: could not load img: %s\n", img_path)
			LOCK.Lock()
			trk["errored"]++
			LOCK.Unlock()
			continue
		}

		fmt.Printf("working on: %s\n", img_path)
		if err := img.proc_img(a, trk); err != nil {
			log.Printf("error processing img: %s\n%s\n", img.src_path, err)
			LOCK.Lock()
			trk["errored"]++
			LOCK.Unlock()
		}

	}

	// recursive case
	if !a.rec {
		return
	}

	for _, file := range files {
		if !file.IsDir() {
			continue
		}

		next_src_dir := filepath.Join(path, file.Name())
		if treeHasImg(next_src_dir, false) {

			next_out_dir := ""
			if a.flat {
				next_out_dir = out_dir
			} else {
				next_out_dir = filepath.Join(out_dir, file.Name())
				if err := os.MkdirAll(next_out_dir, 0755); err != nil {
					log.Fatal(err)
				}
			}

			wg.Add(1)
			go process_files(
				next_src_dir,
				next_out_dir,
				a, trk,
			)

		}
	}
	wg.Done()

}
