package identity

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base32"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

type Params struct {
	Memory      uint32 // RAM usage
	Iterations  uint32 // iterations (CPU time)
	Parallelism uint8  // number of parallel threads
	SaltLength  uint32 // salt length (bytes)
	KeyLength   uint32 // output hash length (bytes)
}

var defaultParams = Params{
	Memory:      64 * 1024,
	Iterations:  3,
	Parallelism: 4, // smaller or equal to machine's cpu cores
	SaltLength:  16,
	KeyLength:   32,
}

func HashPassword(password string) (string, error) {
	// create random salt
	salt := make([]byte, defaultParams.SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	// calculate raw argon2id
	hash := argon2.IDKey(
		[]byte(password),
		salt,
		defaultParams.Iterations,
		defaultParams.Memory,
		defaultParams.Parallelism,
		defaultParams.KeyLength,
	)

	// encode salt and hash to base64
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	// build standard encoded string
	encodedHash := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		defaultParams.Memory,
		defaultParams.Iterations,
		defaultParams.Parallelism,
		b64Salt,
		b64Hash,
	)

	return encodedHash, nil
}

func VerifyPassword(password, encodedHash string) (bool, error) {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return false, errors.New("invalid encoded hash format")
	}
	// check argon algorithm
	if parts[1] != "argon2id" {
		return false, errors.New("unsupported argon2 variant")
	}

	// check version
	var version int
	_, err := fmt.Sscanf(parts[2], "v=%d", &version)
	if err != nil || version != argon2.Version {
		return false, errors.New("incompatible argon2 version")
	}

	// read hardware params (m, t, p)
	var params Params
	_, err = fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &params.Memory, &params.Iterations, &params.Parallelism)
	if err != nil {
		return false, errors.New("invalid argon2 parameters")
	}
	if params.Iterations < 1 || params.Parallelism < 1 {
		return false, errors.New("invalid argon2 parameters")
	}
	if params.Memory > 1<<20 {
		return false, errors.New("argon2 memory parameter too large")
	}

	// decode salt
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, err
	}
	// decode hash
	storedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, err
	}
	if len(salt) == 0 || len(storedHash) == 0 {
		return false, errors.New("invalid encoded hash format")
	}

	params.KeyLength = uint32(len(storedHash))

	// calculate the hash value of given password
	computedHash := argon2.IDKey(
		[]byte(password),
		salt,
		params.Iterations,
		params.Memory,
		params.Parallelism,
		params.KeyLength,
	)

	if subtle.ConstantTimeCompare(storedHash, computedHash) == 1 {
		return true, nil
	}

	return false, nil
}

func NewSecurityStamp() string {
	b := make([]byte, 20)
	rand.Read(b)
	return base32.StdEncoding.EncodeToString(b)
}
