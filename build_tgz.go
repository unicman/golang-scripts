/*
Utility script to generate tar gz archive of specified file.
For now it supports only file paths ... no wildcards or directory inclusion.
*/
package main

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"os"
)

const USAGE = `
Usage:
go run build_tgz.go tgz_path file1_path file2_path ...

Where,
tgz_path    - relative or absolute path of output archive including file name and extension
fileNN_path - relative path of input file. DO NOT specify folder. Sub-folders shall be exactly as per
              relative directory location of the file.
`

func main() {
	argsWithoutProg := os.Args[1:]

	if len(argsWithoutProg) < 2 {
		fmt.Fprintf(os.Stderr, USAGE)
		os.Exit(1)
	}

	tgz_path := argsWithoutProg[0]
	// Files which to include in the tar.gz archive
	files := argsWithoutProg[1:]

	// Create output file
	out, err := os.Create(tgz_path)
	if err != nil {
		log.Fatalln("Error writing archive:", err)
	}
	defer out.Close()

	// Create the archive and write the output to the "out" Writer
	err = createArchive(files, out)
	if err != nil {
		log.Fatalln("Error creating archive:", err)
	}

	fmt.Println("Archive created successfully")
}

func createArchive(files []string, buf io.Writer) error {
	// Create new Writers for gzip and tar
	// These writers are chained. Writing to the tar writer will
	// write to the gzip writer which in turn will write to
	// the "buf" writer
	gw := gzip.NewWriter(buf)
	defer gw.Close()
	tw := tar.NewWriter(gw)
	defer tw.Close()

	// Iterate over files and add them to the tar archive
	for _, file := range files {
		err := addToArchive(tw, file)
		if err != nil {
			return err
		}
	}

	return nil
}

func addToArchive(tw *tar.Writer, filename string) error {
	// Open the file which will be written into the archive
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Get FileInfo about our file providing file size, mode, etc.
	info, err := file.Stat()
	if err != nil {
		return err
	}

	// Create a tar Header from the FileInfo data
	header, err := tar.FileInfoHeader(info, info.Name())
	if err != nil {
		return err
	}

	// Use full path as name (FileInfoHeader only takes the basename)
	// If we don't do this the directory strucuture would
	// not be preserved
	// https://golang.org/src/archive/tar/common.go?#L626
	header.Name = filename

	// Write file header to the tar archive
	err = tw.WriteHeader(header)
	if err != nil {
		return err
	}

	// Copy file content to tar archive
	_, err = io.Copy(tw, file)
	if err != nil {
		return err
	}

	return nil
}
