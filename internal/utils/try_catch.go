package utils

// Exception type
type Exception interface{}

// Block
type Block struct {
	Try     func()
	Catch   func(Exception)
	Finally func()
}

// Throw
func Throw(ex Exception) {
	panic(ex)
}

// Try-catch-finally
func (b Block) Do() {
	if b.Finally != nil {
		defer b.Finally()
	}
	if b.Catch != nil {
		defer func() {
			if r := recover(); r != nil {
				b.Catch(r)
			}
		}()
	}
	b.Try()
}
