package fileconfig

// import (
// 	"os"
// )

// type BlobFileConfig struct {
// 	fileConfig
// }

// func NewBlobConfig(filePath string, data []byte) (FileConfig, error) {
// 	var bc BlobFileConfig
// 	bc.fileConfig.FilePath = filePath
// 	if err := bc.Load(data); err != nil {
// 		return nil, err
// 	}
// 	return &bc, nil
// }

// func (bc *BlobFileConfig) Load(data []byte) error {
// 	var value any
// 	if _, err := os.Stat(bc.fileConfig.FilePath); err == nil {
// 		blob, err := os.ReadFile(bc.fileConfig.FilePath)
// 		if err != nil {
// 			return err
// 		}
// 		value, err = bc.fileConfig.decode(blob)
// 		if err != nil {
// 			return err
// 		}
// 	}

// 	if value != nil {
// 		bc.fileConfig.Buffer = []byte(value.(string))
// 	} else if len(data) > 0 {
// 		bc.fileConfig.Buffer = data
// 	} else {
// 		bc.fileConfig.Buffer = []byte("")
// 	}

// 	return nil
// }

// func (bc *BlobFileConfig) Save() error {
// 	f, err := os.OpenFile(bc.fileConfig.FilePath,
// 		os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.ModePerm)
// 	if err != nil {
// 		return err
// 	}
// 	defer f.Close()

// 	enc, err := bc.fileConfig.encode(string(bc.fileConfig.Buffer))
// 	if err != nil {
// 		return err
// 	}
// 	_, err = f.Write(enc)
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }

// func (bc *BlobFileConfig) Dump() ([]byte, error) {
// 	_, err := os.Stat(bc.fileConfig.FilePath)
// 	if err != nil {
// 		return nil, err
// 	}

// 	blob, err := os.ReadFile(bc.fileConfig.FilePath)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return blob, nil
// }

// func (bc *BlobFileConfig) Buffer() []byte {
// 	return bc.fileConfig.Buffer
// }
