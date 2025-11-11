package query

type Fuzziness string

const (
	NoFuzziness   Fuzziness = "NONE"
	FuzzinessAuto Fuzziness = "AUTO"
	Fuzziness0    Fuzziness = "0"
	Fuzziness1    Fuzziness = "1"
	Fuzziness2    Fuzziness = "2"
)

var SupportedFuzziness = map[Fuzziness]bool{
	FuzzinessAuto: true,
	Fuzziness0:    true,
	Fuzziness1:    true,
	Fuzziness2:    true,
}
