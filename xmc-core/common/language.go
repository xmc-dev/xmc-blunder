package common

import (
	"fmt"
)

// ErrInvalidLanguage is returned when there is no programming language in the system that matches the extension
type ErrInvalidLanguage string

type Language string

const (
	LangC   Language = "c"
	LangCPP          = "cpp"
	//	LangJava                = "java"
	//	LangPython2             = "py"
	//	LangPython3             = "py3"
	//	LangHaskell             = "hs"
	//	LangRuby                = "rb"
	//	LangCommonLISP          = "lisp"
	//	LangPascal              = "pas"
	LangGo = "go"
	end
)

func (eil ErrInvalidLanguage) Error() string {
	return fmt.Sprintf("invalid file extension %s for programming language", string(eil))
}

// FileExtToLanguage returns a programming language code from a file extension.
//
// The actual string value of the constant must also be a valid extension.
func FileExtToLanguage(ext string) (Language, error) {
	switch {
	case ext == "c":
		return LangC, nil
	case ext == "cpp" || ext == "cxx" || ext == "C":
		return LangCPP, nil
		//	case ext == "java":
		//		return LangJava, nil
		//	case ext == "py":
		//		return LangPython2, nil
		//	case ext == "py3":
		//		return LangPython3, nil
		//	case ext == "hs":
		//		return LangHaskell, nil
		//	case ext == "rb":
		//		return LangRuby, nil
		//	case ext == "lisp":
		//		return LangCommonLISP, nil
		//	case ext == "pas":
		//		return LangPascal, nil
	case ext == "go":
		return LangGo, nil
	default:
		return "", ErrInvalidLanguage(ext)
	}
}

func IsValidLanguage(lang string) bool {
	_, err := FileExtToLanguage(lang)
	return err == nil
}
