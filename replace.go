package wellington

// Replace iterates through the list of substrings to
// cut or replace, adjusting for shift of the output
// buffer as a result of these ops.
func (p *Parser) Replace() {

	for _, c := range p.Chop {
		begin := c.Start - p.shift
		end := c.End - p.shift
		if begin < 0 {
			continue
		}
		p.shift += end - begin
		p.shift -= len(c.Value)
		suf := append(c.Value, p.Output[end:]...)

		p.Output = append(p.Output[:begin], suf...)
	}
}

// Mark segments of the input string for future deletion.
func (p *Parser) Mark(start, end int, val string) {
	// fmt.Println("Mark:", string(p.Input[start:end]), "~>", val)
	p.Chop = append(p.Chop, Replace{start, end, []byte(val)})
}
