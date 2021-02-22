// Copyright (c) 2020, 2021, Oracle and/or its affiliates.
// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.
// Originally written in Java by Mark Nelson translated to Go by Mark Nelson and Mike Cico.

package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// This program will accept a list of files and directories and scan all of the files found therin to make sure that
// they have the correct Oracle copyright header and UPL license headers.
//
// Internally, we manage a list of file extensions and relative file/directory names to ignore.  We also load a list
// of ignore paths from the working directory of the program containing a list of paths relative to that working dir
// to explicitly ignore.

const (
	// ignoreFileName is the name of the special file that contains a list of files to ignore
	ignoreFileName = "ignore_copyright_check.txt"

	// maxLines is the maximum number of lines to read in a file before giving up
	maxLines = 5
)

var (
	// filesToSkip is a list of well-known filenames to skip while scanning, relative to the directory being scanned
	filesToSkip = []string{
			".gitlab-ci.yml",
			"go.mod",
			"go.sum",
			"LICENSE",
			"LICENSE.txt",
			"THIRD_PARTY_LICENSES.txt",
			"coverage.html",
			"clair-scanner",
			".DS_Store",
	}

	// directoriesToShip is a list of well-known (sub)directories to skip while scanning, relative to the working
	// directory being scanned
	directoriesToSkip = []string{
		".git",
		"out",
		"bin",
		".github",
		".idea",
		".settings",
		"thirdparty_licenses",
		"vendor",
		"_output",
		"_gen", "target",
		"node_modules",
	}

	// extensionsToSkip is a list of well-known file extensions that we will skip while scanning, including
	// binary files and file types that do not support comments (like json)
	extensionsToSkip = []string{
		".json",
		".png",
		".csv",
		".ico",
		".md",
		".jpeg",
		".jpg",
		".log",
		"-test-result.xml",
		".woff",
		".woff2",
		".ttf",
		".min.js",
		".min.css",
		".map",
		".cov",
		".iml",
	}

	// copyrightRegex is the regular expression for recognizing correctly formatted copyright statements
	// Explanation of the regular expression
	// -------------------------------------
	// ^                           matches start of the line
	// (#|\/\/|<!--|\/\*)          matches either a # character, or two / characters or the literal string "<!--", or "/*"
	// Copyright                   matches the literal string " Copyright "
	// \([cC]\)                    matches "(c)" or "(C)"
	// ([1-2][0-9][0-9][0-9], )    matches a year in the range 1000-2999 followed by a comma and a space
	// ?([1-2][0-9][0-9][0-9], )   matches an OPTIONAL second year in the range 1000-2999 followed by a comma and a space
	// Oracle ... affiliates       matches that literal string
	// (\.|\. -->|\. \*\/|\. --%>) matches "." or ". -->" or ". */"
	// $                           matches the end of the line
	// the correct copyright line looks like this:
	// Copyright (c) 2020, Oracle and/or its affiliates.
	copyrightRegex = regexp.MustCompile(`^(#|\/\/|<!--|\/\*|<%--) Copyright \([cC]\) ([1-2][0-9][0-9][0-9], )?([1-2][0-9][0-9][0-9], )Oracle and\/or its affiliates(\.|\. -->|\. \*\/|\. --%>)$`)

	// uplRegex is the regular express for recognizing correctly formatted UPL license headers
	// Explanation of the regular expression
	// -------------------------------------
	// ^                           matches start of the line
	// (#|\/\/|<!--|\/\*|<%--)     matches either a # character, or two / characters or the literal string "<!--", "/*" or "<%--"
	// Licensed ... licenses\\/upl matches that literal string
	// (\.|\. -->|\. \*\/|\. --%>) matches "." or ". -->" or ". */" or ". --%>"
	// $                           matches the end of the line
	// the correct copyright line looks like this:
	// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.
	uplRegex = regexp.MustCompile(`^(#|\/\/|<!--|\/\*|<%--) Licensed under the Universal Permissive License v 1\.0 as shown at https:\/\/oss\.oracle\.com\/licenses\/upl(\.|\. -->|\. \*\/|\. --%>)$`)

	// filesWithErrors List to track files that failed the check
	filesWithErrors = []string{}
	// numFilesAnalyzed Total number of files analyzed
	numFilesAnalyzed int32 = 0
	// numFilesSkipped Total number of files skipped
	numFilesSkipped int32 = 0
	// numDirectoriesSkipped Total number of directories skipped
	numDirectoriesSkipped int32 = 0

	// filesToIgnore Files to ignore
	filesToIgnore = []string{}

	// directoriesToIgnore Directories to ignore
	directoriesToIgnore = []string{}
)

func main() {
	args := os.Args
	if len(args) < 2 {
		printUsage()
		return
	}

	if err := loadIgnoreFile(); err != nil {
		fmt.Print("Error updating ingore files list: %v", err)
		os.Exit(1)
	}

	// Arguments are a list of directories and/or files.  Iterate through each one and
	// - if it's a file,scan it
	// - if it's a dir, walk it and scan it recursively
	for _, arg := range args[1:] {
		fmt.Println(fmt.Sprintf("Copyright scanner scanning %s", arg))
		argInfo, err := os.Stat(arg)
		if err != nil {
			fmt.Printf("Error getting file info for %s: %v", arg, err.Error())
			os.Exit(1)
		}
		if argInfo.IsDir() {
			err = filepath.Walk(arg, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if info.IsDir() {
					if skipOrIgnoreDir(info.Name(), path) {
						return filepath.SkipDir
					}
					return nil
				}
				err = checkFile(path, info)
				if err != nil {
					return err
				}
				return nil
			})
		} else {
			err = checkFile(arg, argInfo)
		}
		if err != nil {
			fmt.Printf("Error processing %s: %v", arg, err.Error())
			os.Exit(1)
		}
	}
	printScanReport()
}

// skipOrIgnoreDir Returns true if a directory matches the skip or ignore lists
func skipOrIgnoreDir(relativeName string, path string) bool {
	if contains(directoriesToSkip, relativeName) || contains(directoriesToIgnore, path) {
		fmt.Println(fmt.Sprintf("Ignoring %s", path))
		numDirectoriesSkipped++
		return true
	}
	return false
}

// contains Search a list of strings for a value
func contains(strings []string, value string) bool {
	for i, _ := range strings {
		if value == strings[i] {
			return true
		}
	}
	return false
}

// checkFile Scans the specified file if it does not match the ignore criteria
func checkFile(path string, info os.FileInfo) error {
	// Ignore the file if
	// - the extension matches one in the global set of ignored extensions
	// - the name matches one in the global set of ignored relative file names
	// - it is in the global ignores list read from disk
	if contains(extensionsToSkip, filepath.Ext(info.Name())) ||
		contains(filesToSkip, info.Name()) ||
		contains(filesToIgnore, path)  {
		numFilesSkipped++
		fmt.Println(fmt.Sprintf("Ignoring %s", path))
		return nil
	}

	numFilesAnalyzed++
	copyrightFound, licenseFound, err := hasValidCopyright(path)
	if err != nil {
		return err
	}
	if !copyrightFound || !licenseFound {
		// append to failed list
		filesWithErrors = append(filesWithErrors, path)
	}
	return nil
}

// hasValidCopyright returns true if the file has a valid/correct copyright notice
func hasValidCopyright(path string) (foundCopyright bool, foundLicense bool, err error) {
	file, err := os.Open(path)
	if err != nil {
		return false, false, err
	}
	reader := bufio.NewScanner(file)
	reader.Split(bufio.ScanLines)
	defer file.Close()

	linesRead := 0
	for reader.Scan() && linesRead < maxLines {
		line := reader.Text()
		if copyrightRegex.MatchString(line) {
			foundCopyright = true
		}
		if uplRegex.MatchString(line) {
			foundLicense = true
		}
		if foundCopyright && foundLicense {
			break
		}
		linesRead++
	}
	return foundCopyright, foundLicense, nil
}

func printScanReport() {
	fmt.Printf("\nResults of scan:\n\tFiles analyzed: %d\n\tFiles with error: %d\n\tFiles skipped: %d\n\tDirectories skipped: %d\n",
	 numFilesAnalyzed, len(filesWithErrors), numFilesSkipped, numDirectoriesSkipped);

	if len(filesWithErrors) > 0 {
		fmt.Println("The following files have errors:\n")
		for _, fileWithError := range filesWithErrors {
			fmt.Println(fileWithError)
		}

		fmt.Println("\nExamples of valid comments:")
		fmt.Println("With forward slash (Java-style):")
		fmt.Println("// Copyright (c) 2021, Oracle and/or its affiliates.")
		fmt.Println("// Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.")
		fmt.Println("With dash (For SQL files for example):")
		fmt.Println("-- Copyright (c) 2021, Oracle and/or its affiliates.")
		fmt.Println("-- Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.")
		fmt.Println("XML comments:")
		fmt.Println("<!-- Copyright (c) 2021, Oracle and/or its affiliates. -->")
		fmt.Println("<!-- Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl. -->")
		fmt.Println("With #:")
		fmt.Println("# Copyright (c) 2021, Oracle and/or its affiliates.")
		fmt.Println("# Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl.")
	}
}

func loadIgnoreFile() error {
	ignoreFile, err := os.Open(ignoreFileName)
	if err != nil {
		return err
	}
	reader := bufio.NewScanner(ignoreFile)
	reader.Split(bufio.ScanLines)
	defer ignoreFile.Close()

	// ignoreFileList Contents of ignore file
	ignoreFileList := []string{}

	for reader.Scan() {
		line := strings.TrimSpace(reader.Text())
		// skip empty lines - otherwise the code below will end up skipping entire
		if len(line) == 0 {
			continue
		}
		// ignore lines starting with "#"
		if strings.HasPrefix(line, "#") {
			continue
		}
		ignoreFileList = append(ignoreFileList, line)
	}

	for _, ignoreLine := range ignoreFileList {
		info, err := os.Stat(ignoreLine)
		if err != nil {
			continue
		}
		if info.IsDir() {
			// if the path points to an existing directory, add it to directories to ignore
			directoriesToIgnore = append(directoriesToIgnore, ignoreLine)
		} else {
			filesToIgnore = append(filesToIgnore, ignoreLine)
		}
	}

	fmt.Printf("Files to ignore: %v\n", filesToIgnore)
	fmt.Printf("Directories to ignore: %v\n", directoriesToIgnore)
	return nil
}

func printUsage() {
	fmt.Println("Copyright scanner")
	fmt.Println("Specify a list of files and or directories to scan")
}