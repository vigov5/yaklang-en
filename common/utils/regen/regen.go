package regen

import (
	"regexp/syntax"

	"github.com/yaklang/yaklang/common/log"

	"github.com/pkg/errors"
)

type CaptureGroupHandler func(index int, name string, group *syntax.Regexp, generator Generator, args *GeneratorArgs) []string

type GeneratorArgs struct {
	Flags               syntax.Flags
	CaptureGroupHandler CaptureGroupHandler
}

func (a *GeneratorArgs) initialize() error {
	if (a.Flags&syntax.UnicodeGroups) == syntax.UnicodeGroups && (a.Flags&syntax.Perl) != syntax.Perl {
		return errors.New("UnicodeGroups not supported")
	}

	if a.CaptureGroupHandler == nil {
		a.CaptureGroupHandler = defaultCaptureGroupHandler
	}

	return nil
}

type Generator interface {
	Generate() []string
	String() string
	CheckVisible(str string) bool
}

// . Generate Generate all matching strings based on the regular expression and return the generated string slice and error
// . For some metacharacters that may match multiple times:
// *     : it will only generate a string that matches 0 or 1 times
// +: will only generate strings matching 1 or 2 times.
// {n,m} : then it will generate a string that matches n to m times
// {n,}  : will only generate matches n times. Or string n+1 times
// Example:
// ```
// regen.Generate("[a-z]+") // a-z single letter, aa-zz two letters
// ```
func Generate(pattern string) ([]string, error) {
	generator, err := NewGenerator(pattern, &GeneratorArgs{
		Flags: syntax.Perl,
	})
	if err != nil {
		return []string{""}, err
	}
	return generator.Generate(), nil
}

// GenerateOne generates a matching string based on the regular expression and returns the generated string and error
// Example:
// ```
// regen.GenerateOne("[a-z]") // a-z
// regen.GenerateOne("^(13[0-9]|14[57]|15[0-9]|18[0-9])\d{8}$") // . Generate a mobile phone number
// ```
func GenerateOne(pattern string) (string, error) {
	generator, err := NewGeneratorOne(pattern, &GeneratorArgs{
		Flags: syntax.Perl,
	})
	if err != nil {
		return "", err
	}
	return generator.Generate()[0], nil
}

// GenerateVisibleOne Generates a matching string (all visible characters) according to the regular expression, returns the generated string and a random letter in the error
// Example:
// ```
// regen.GenerateVisibleOne("[a-z]") // a-z
// regen.GenerateVisibleOne("^(13[0-9]|14[57]|15[0-9]|18[0-9])\d{8}$") // . Generate a mobile phone number
// ```
func GenerateVisibleOne(pattern string) (string, error) {
	generator, err := NewGeneratorVisibleOne(pattern, &GeneratorArgs{
		Flags: syntax.Perl,
	})
	if err != nil {
		return "", err
	}
	generated := generator.Generate()[0]
	if len(generated) > 0 {
		if !generator.CheckVisible(generated) {
			log.Warnf("pattern %s,res [%s] is not visible one", pattern, generated)
		}
	}
	return generated, nil
}

// MustGenerate Generates all matching characters according to the regular expression String, if the generation fails, it will crash and return the generated string slice
// . For some metacharacters that may match multiple times:
// *     : it will only generate a string that matches 0 or 1 times
// +: will only generate strings matching 1 or 2 times.
// {n,m} : then it will generate a string that matches n to m times
// {n,}  : will only generate matches n times. Or string n+1 times
// Example:
// ```
// regen.MustGenerate("[a-z]+") // a-z single letter, aa-zz two letters
// ```
func MustGenerate(pattern string) []string {
	generator, err := NewGenerator(pattern, &GeneratorArgs{
		Flags: syntax.Perl,
	})
	if err != nil {
		panic(err)
	}
	return generator.Generate()
}

// MustGenerateOne Generate a matching string based on the regular expression. If the generation fails, it will crash and return the generated string
// Example:
// ```
// regen.MustGenerateOne("[a-z]") // a-z
// regen.MustGenerateOne("^(13[0-9]|14[57]|15[0-9]|18[0-9])\d{8}$") // . Generate a mobile phone number
// ```
func MustGenerateOne(pattern string) string {
	generator, err := NewGeneratorOne(pattern, &GeneratorArgs{
		Flags: syntax.Perl,
	})
	if err != nil {
		panic(err)
	}
	return generator.Generate()[0]
}

// . MustGenerateVisibleOne generates a matching string (all visible characters) based on the regular expression. If the generation fails, it will crash and return the generated string
// Example:
// ```
// regen.MustGenerateVisibleOne("[a-z]") // a-z
// regen.MustGenerateVisibleOne("^(13[0-9]|14[57]|15[0-9]|18[0-9])\d{8}$") // . Generate a mobile phone number
// ```
func MustGenerateVisibleOne(pattern string) string {
	generator, err := NewGeneratorVisibleOne(pattern, &GeneratorArgs{
		Flags: syntax.Perl,
	})
	if err != nil {
		panic(err)
	}
	generated := generator.Generate()[0]
	if len(generated) > 0 {
		if !generator.CheckVisible(generated) {
			log.Warnf("pattern %s,res [%s] is not visible one", pattern, generated)
		}
	}
	return generated
}

func NewGenerator(pattern string, inputArgs *GeneratorArgs) (generator Generator, err error) {
	args := GeneratorArgs{}

	// Copy inputArgs so the caller can't change them.
	if inputArgs != nil {
		args = *inputArgs
	}
	if err = args.initialize(); err != nil {
		return nil, err
	}

	var regexp *syntax.Regexp
	regexp, err = syntax.Parse(pattern, args.Flags)
	if err != nil {
		return
	}

	var gen *internalGenerator
	gen, err = newGenerator(regexp, &args)
	if err != nil {
		return
	}

	return gen, nil
}

func NewGeneratorOne(pattern string, inputArgs *GeneratorArgs) (geneator Generator, err error) {
	args := GeneratorArgs{}

	// Copy inputArgs so the caller can't change them.
	if inputArgs != nil {
		args = *inputArgs
	}
	if err = args.initialize(); err != nil {
		return nil, err
	}

	var regexp *syntax.Regexp
	regexp, err = syntax.Parse(pattern, args.Flags)
	if err != nil {
		return
	}

	var gen *internalGenerator
	gen, err = newGeneratorOne(regexp, &args)
	if err != nil {
		return
	}

	return gen, nil
}

func NewGeneratorVisibleOne(pattern string, inputArgs *GeneratorArgs) (geneator Generator, err error) {
	args := GeneratorArgs{}

	// Copy inputArgs so the caller can't change them.
	if inputArgs != nil {
		args = *inputArgs
	}
	if err = args.initialize(); err != nil {
		return nil, err
	}

	var regexp *syntax.Regexp
	regexp, err = syntax.Parse(pattern, args.Flags)
	if err != nil {
		return
	}

	var gen *internalGenerator
	gen, err = newGeneratorVisibleOne(regexp, &args)
	if err != nil {
		return
	}

	return gen, nil
}
