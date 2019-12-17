package main

/* This program computes marginal, conditional, and joint probabilities from a
 * read matrix of indicator variables. The input is provided on stdin. The first
 * row must be the column names. The first column must be labeled 'read' and must
 * give the names of the reads. All the other columns must be 0 or 1 indicator
 * variable columns. */

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
)

type Args struct {
	Limit        int
	Marginals    string
	Conditionals string
	Joints       string
}

var args = Args{}

func init() {
	log.SetFlags(0)
	flag.StringVar(&args.Marginals, "marginals", "", "file to write marginal probabilities to")
	flag.StringVar(&args.Joints, "joints", "", "file to write joint probabilities to")
	flag.StringVar(&args.Conditionals, "conditionals", "", "file to write conditional probabilities to")
	flag.IntVar(&args.Limit, "limit", 0, "limit the number of lines of stdin to consider (default = 0 = unlimited)")

	flag.Usage = func() {
		log.Println("usage: matrixprobs [options] < matrix.tsv")
		flag.PrintDefaults()
	}
}

func main() {
	flag.Parse()

	if args.Marginals == "" && args.Joints == "" && args.Conditionals == "" {
		log.Println("Must specify at least one of -marginals, -joints, and/or -conditionals")
		flag.Usage()
		os.Exit(1)
	}

	var fieldNames []string
	var row []int
	var marginals []int
	var joints [][]int
	calcJoints := args.Joints != "" || args.Conditionals != ""
	numFields := 0

	var marFp *os.File
	var jointFp *os.File
	var condFp *os.File

	// Get the output descriptors ready now so we fail early
	if args.Marginals != "" {
		var err error
		marFp, err = os.Create(args.Marginals)
		if err != nil {
			log.Fatalf("failed to open marginals file '%s': %v\n", args.Marginals, err)
		}
	}
	if args.Joints != "" {
		var err error
		jointFp, err = os.Create(args.Joints)
		if err != nil {
			log.Fatalf("failed to open joints file '%s': %v\n", args.Joints, err)
		}
	}
	if args.Conditionals != "" {
		var err error
		condFp, err = os.Create(args.Conditionals)
		if err != nil {
			log.Fatalf("failed to open conditionals file '%s': %v\n", args.Conditionals, err)
		}
	}

	scanner := bufio.NewScanner(os.Stdin)
	lineNum := 0
	numReads := 0
	for scanner.Scan() {
		if args.Limit > 0 && lineNum > args.Limit {
			break
		}
		line := scanner.Text()
		lineNum++
		fields := strings.Split(line, "\t")
		if lineNum == 1 {
			// This should be the header line
			if len(fields) < 2 {
				log.Fatalln("too few fields")
			}
			if fields[0] != "read" {
				log.Fatalln("first field should be named 'read'")
			}
			fieldNames = fields[1:]
			numFields = len(fieldNames)
			log.Println("number of fields:", numFields)
			// Allocate data structures we'll need
			row = make([]int, numFields)
			marginals = make([]int, numFields)
			joints = make([][]int, numFields)
			for i := 0; i < numFields; i++ {
				joints[i] = make([]int, numFields)
			}
		} else {
			// This should be a line naming the read and giving the values of the indicator variables
			if len(fields) != numFields+1 {
				log.Fatalf("expected line %d to have %d fields\n", lineNum, numFields+1)
			}
			// parse the row and tally marginal counts
			for i, _ := range fieldNames {
				val := fields[i+1]
				if val == "0" {
					row[i] = 0
				} else if val == "1" {
					row[i] = 1
				} else {
					log.Fatalf("invalid value '%s' on line %d\n", val, lineNum)
				}
				marginals[i] += row[i]
			}
			// joint counts
			if calcJoints {
				for i, _ := range fieldNames {
					for j, _ := range fieldNames {
						if row[i]*row[j] == 1 {
							joints[i][j] += 1
						}
					}
				}
			}
			numReads += 1
		}
	}
	// Print the marginals
	if args.Marginals != "" {
		for i, name := range fieldNames {
			mar := float64(marginals[i]) / float64(numReads)
			fmt.Fprintf(marFp, "P( %s ) = %0.6f ; %d\n", name, mar, marginals[i])
		}
	}
	// Print the joint probabilities
	if args.Joints != "" {
		for i, iName := range fieldNames {
			for j, jName := range fieldNames {
				// P(A^B) = joint/numReads
				jointProb := float64(joints[i][j]) / float64(numReads)
				fmt.Fprintf(jointFp, "P( %s , %s ) = %0.8f ; %d\n", iName, jName, jointProb, joints[i][j])
			}
		}
	}
	// Print the conditional probabilities
	if args.Conditionals != "" {
		for i, iName := range fieldNames {
			for j, jName := range fieldNames {
				// P(A^B) = joint/numReads
				// P(A | B) = P(A^B) / P(B)
				if marginals[i] == 0 {
					fmt.Fprintf(condFp, "P( %s | %s ) = NaN ; %d , %d\n", jName, iName, joints[i][j], marginals[i])
					continue
				}
				jointProb := float64(joints[i][j]) / float64(numReads)
				mar := float64(marginals[i]) / float64(numReads)
				condProb := jointProb / mar
				fmt.Fprintf(condFp, "P( %s | %s ) = %0.8f ; %d , %d\n", jName, iName, condProb, joints[i][j], marginals[i])
			}
		}
	}
}
