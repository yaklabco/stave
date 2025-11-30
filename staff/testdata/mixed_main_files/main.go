// some points here: this is the main package just like the stavefiles.  This one
// has a main function, which could conflict with the generated main function we
// make (but clearly shouldn't cause problems). Finally, there's a duplicate function name here.

package main

func main() {}

func Build() {}
