package main

/* TODO
* Use separate file for max attempts counter
	* Keep track of when it was updated in the encoded json creds file
	* If the file was deleted or changed by a third-party (?), lock the program
* Take all input from stdin, not as command-line args; clear screen afterwards
*/

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"runtime"
	"strings"

	"golang.org/x/term"
)

type creds map[string]string

type credsSite map[string]creds

type credentialsData struct {
	Master      string
	Attempts    int
	MaxAttempts int
	Sites       map[string]credsSite
}

func loadCreds(master, credsFile string) (*credentialsData, error) {
	allCreds := &credentialsData{}
	masterSum := sha256.Sum256([]byte(master))
	block, err := aes.NewCipher(masterSum[:])
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	ciphertext, err := ioutil.ReadFile(credsFile)
	nonce := ciphertext[:gcm.NonceSize()]
	ciphertext = ciphertext[gcm.NonceSize():]
	credsJSON, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil,
			fmt.Errorf(
				"error reading creds file: %s (most likely incorrect master password)",
				err.Error())
	}
	if err := json.Unmarshal(credsJSON, allCreds); err != nil {
		return nil, fmt.Errorf("error reading creds file: %s", err.Error())
	}
	return allCreds, nil
}

// List lists the credentials of the given site (and main credential if given)
func (cd *credentialsData) List(site, mainCredential string) error {
	if site == "" {
		// List every site if no site is provided
		for site := range cd.Sites {
			fmt.Println(site)
		}
		return nil
	}
	cs, ok := cd.Sites[site]
	if !ok {
		return fmt.Errorf("invalid site")
	}
	if mainCredential == "" {
		// List every main credential if no main credential is provided
		for mainCred := range cs {
			fmt.Println(mainCred)
		}
		return nil
	}
	c, ok := cs[mainCredential]
	if !ok {
		return fmt.Errorf("invalid main credential")
	}
	// List credential set
	for field, value := range c {
		fmt.Printf("%s | %s\n", field, value)
	}
	return nil
}

// Set updates a credential set
func (cd *credentialsData) Set(site, mainCredential string, fields creds) error {
	delete(fields, "site")
	if site == "" {
		return fmt.Errorf("must provide site")
	} else if mainCredential == "" {
		return fmt.Errorf("must provide main credential")
	} else if len(fields) == 0 {
		return fmt.Errorf("must provide fields (key:value pairs)")
	}
	cs, ok := cd.Sites[site]
	if !ok {
		cs = make(credsSite)
		cd.Sites[site] = cs
	}
	c, ok := cs[mainCredential]
	if !ok {
		c = make(creds)
		cs[mainCredential] = c
	}
	for field, value := range fields {
		c[field] = value
	}
	return nil
}

// Delete deletes a given credential set or site itself
// "fields" argument unneeded right now
func (cd *credentialsData) Delete(site, mainCredential string, fields creds) error {
	delete(fields, "site")
	if site == "" {
		return fmt.Errorf("must provide site")
	}
	cs, ok := cd.Sites[site]
	if !ok {
		return fmt.Errorf("invalid site")
	}
	c, ok := cs[mainCredential]
	if !ok && mainCredential != "" {
		return fmt.Errorf("invalid main credential")
	}
	if mainCredential == "" {
		// Delete entire site if main credential isn't given
		delete(cd.Sites, site)
	} else if len(fields) == 0 || true {
		/* ALERT: With just len check, this block will never be reached, code won't work right */
		// Delete credential set if fields aren't provided
		delete(cs, mainCredential)
	}
	if false {
		// Delete fields from credential set
		for field := range fields {
			if _, ok := c[field]; !ok {
				log.Printf("invalid field: %s\n", field)
			} else {
				delete(c, field)
			}
		}
	}
	return nil
}

func (cd *credentialsData) printAll() error {
	j, err := json.MarshalIndent(cd, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(j))
	return nil
}

func (cd *credentialsData) exportData(credsFile string) error {
	masterSum := sha256.Sum256([]byte(cd.Master))
	block, err := aes.NewCipher(masterSum[:])
	if err != nil {
		return fmt.Errorf("error writing creds file: %s", err.Error())
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("error writing creds file: %s", err.Error())
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return fmt.Errorf("error writing creds file: %s", err)
	}
	credsJSON, err := json.Marshal(cd)
	if err != nil {
		return fmt.Errorf("error writing creds file: %s", err)
	}
	ciphertext := gcm.Seal(nonce, nonce, credsJSON, nil)
	if err := ioutil.WriteFile(credsFile, ciphertext, 0777); err != nil {
		return fmt.Errorf("error writing creds file: %s", err)
	}
	return nil
}

func main() {
	log.SetFlags(0)
	n := flag.Bool(
		"new",
		false,
		"(unimplemented) New credential set for the given site (site created if nonexistent)",
	)
	l := flag.Bool(
		"list",
		false,
		"Lists the given credential set, site, or, all available sites",
	)
	s := flag.Bool(
		"set",
		false,
		"Creates/Sets site, credential sets, and/or credential fields for the given site (must have main credential)",
	)
	// Variable not needed because it is not used in program, only used to see
	// whether it was passed as a flag or not
	// Only this one isn't assigned to a variable because it is expected to be
	// used the least, therefore, if no other flag is given, it has to be this
	_ = flag.Bool(
		"delete",
		false,
		"Deletes the credential set or site",
	)
	setMaster := flag.String(
		"set-master",
		"",
		"Sets new master password",
	)
	setMaxAttempts := flag.Int(
		"set-attempts",
		10,
		"Sets new max attempts at password",
	)
	flag.Usage = usage
	flag.Parse()

	// See which flags have actually been passed
	var flagPassed string
	flag.Visit(func(f *flag.Flag) {
		if flagPassed != "" {
			log.Fatal("only one flag can be passed")
		}
		flagPassed = f.Name
	})
	if flagPassed == "" {
		log.Fatal("no flags passed")
	}
	if flagPassed == "new" {
		log.Fatal("flag \"-new\" unimplemented")
	}

	// Securely get the master password from the user
	print("Master password: ")
	bMaster, err := term.ReadPassword(int(os.Stdout.Fd()))
	fmt.Println()
	if err != nil {
		log.Fatal(err)
	}
	master := string(bMaster)

	// Get the credentials from the credentials file
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		log.Fatal("error getting credentials file directory")
	}
	credsFile := path.Join(path.Dir(thisFile), "creds.dat")
	allCreds, err := loadCreds(master, credsFile)
	if err != nil {
		log.Fatal(err)
	}

	// Set the master password or max attempts if given
	if flagPassed == "set-master" {
		if *setMaster == "" {
			log.Fatal("must provide new master password")
		}
		allCreds.Master = *setMaster
		if err := allCreds.exportData(credsFile); err != nil {
			log.Fatal(err)
		}
		return
	} else if flagPassed == "set-attempts" {
		log.Fatal("-set-attempts not implemented")
		println(*setMaxAttempts)
	}

	// Parse the arguments passed
	args := flag.Args()
	if *l {
		combo := strings.Join(args, " ")
		if combo == "print all" {
			allCreds.printAll()
			return
		}
	}
	var mainCred string
	passedCreds := make(creds)
	for _, arg := range args {
		// Get the parts of the argument; split to k:v pair
		parts := strings.SplitN(arg, ":", 2)
		if len(parts) != 2 {
			log.Fatalf("invalid argument: %s (must be in format key:value)", arg)
		}
		// Do a bit of cleaning and checking
		k, v := strings.ToLower(parts[0]), parts[1]
		if k == "" {
			log.Fatalf("%s: must provide key", arg)
		} else if v == "" {
			log.Fatalf("%s: must provide value", v)
		}
		// Set main cred
		if k[0] == '*' {
			if mainCred != "" {
				log.Fatal("can only have one main credential")
			}
			if len(k) == 1 {
				log.Fatal("invalid main credential")
			}
			k = k[1:]
			if k == "password" || k == "site" {
				log.Fatalf("cannot use %s as main cred", k)
			}
			mainCred = v
		}
		// Make sure cred is only passed once
		if val, ok := passedCreds[k]; ok {
			log.Fatalf("duplicate arguments %s:%s and %s", k, val, arg)
		}
		passedCreds[k] = v
	}
	// Set the main cred if it wasn't set before
	if mainCred == "" {
		if email, ok := passedCreds["email"]; ok {
			mainCred = email
		} else if username, ok := passedCreds["username"]; ok {
			mainCred = username
		} else {
			for k, v := range passedCreds {
				if k != "password" && k != "site" {
					mainCred = v
					break
				}
			}
			if *s && mainCred == "" {
				log.Fatal("must provide more than one credential")
			}
		}
	}

	// Run based on flag passed
	if *l {
		if err := allCreds.List(passedCreds["site"], mainCred); err != nil {
			log.Fatal(err)
		}
		return
	} else if *s {
		if err := allCreds.Set(passedCreds["site"], mainCred, passedCreds); err != nil {
			log.Fatal(err)
		}
	} else if *n {
	} else {
		if err := allCreds.Delete(passedCreds["site"], mainCred, passedCreds); err != nil {
			log.Fatal(err)
		}
	}

	if err := allCreds.exportData(credsFile); err != nil {
		log.Fatal(err)
	}
}

func usage() {
	var stmt string
	stmt += "Usage of pwdstore:\n"
	stmt += "\tFormat for all arguments is \"key:value\".\n"
	stmt += "\tIn order to specify the main credential, prefix the key with a '*'.\n"
	stmt += "\tIf a main credential isn't specified, keys equal to \"email\" then \"username\" take preference.\n"
	stmt += "\tAfterwards, a random key is chosen. It can never be a field with the key \"password\" or \"site\".\n"
	stmt += "\tA site must always be specified by passing a key:value pair with the key equal to \"site\".\n"
	stmt += "  -delete\n\tDeletes the credential set or site\n"
	stmt += "\tAny key value can be given for main credential (besides \"site\" and \"password\"); no need for *\n"
	stmt += "  -list\n\tLists the given credential set, site, or all sites\n"
	stmt += "\tAny key value can be given for main credential (besides \"site\" and \"password\"); no need for *\n"
	stmt += "  -new\n\t(unimplemented) New credential set for the given site (site created if nonexistent)\n"
	stmt += "  -set\n\tCreates/Sets site, credential sets, and/or credential fields for the given site (must have main credential)\n"
	stmt += "  -set-attempts\n\tSets new max attempts at password (default 10)\n"
	stmt += "  -set-master string\n\tSets new master password\n"
	log.Println(stmt)
}
