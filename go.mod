go 1.20

module golang.a2z.com/SpaceNeedleHCLWriter

require (
	github.com/spf13/pflag v0.0.0-00010101000000-000000000000
	github.com/vmihailenco/msgpack/v4 v4.0.0-00010101000000-000000000000
	golang.org/x/crypto v0.0.0-00010101000000-000000000000
	golang.org/x/term v0.0.0-00010101000000-000000000000
	golang.org/x/text v0.0.0-00010101000000-000000000000
	github.com/agext/levenshtein v0.0.0-00010101000000-000000000000
	github.com/hashicorp/hcl/v2 v2.0.0-00010101000000-000000000000
	github.com/vmihailenco/tagparser v0.0.0-00010101000000-000000000000
	golang.org/x/net v0.0.0-00010101000000-000000000000
	github.com/zclconf/go-cty/cty v0.0.0-00010101000000-000000000000
	github.com/zclconf/go-cty-debug/ctydebug v0.0.0-00010101000000-000000000000
	github.com/google/go-cmp/cmp v0.0.0-00010101000000-000000000000
	github.com/apparentlymart/go-textseg/v13 v13.0.0-00010101000000-000000000000
	golang.org/x/sys v0.0.0-00010101000000-000000000000
	github.com/mitchellh/go-wordwrap v0.0.0-00010101000000-000000000000
)

replace github.com/spf13/pflag v0.0.0-00010101000000-000000000000 => ./build/private/bgospace/Go3p-Github-Spf13-Pflag/src/github.com/spf13/pflag

replace github.com/vmihailenco/msgpack/v4 v4.0.0-00010101000000-000000000000 => ./build/private/bgospace/Go3p-Github-Vmihailenco-Msgpack-V4/src/github.com/vmihailenco/msgpack/v4

replace golang.org/x/crypto v0.0.0-00010101000000-000000000000 => ./build/private/bgospace/Go3p-Golang-X-Crypto/src/golang.org/x/crypto

replace golang.org/x/term v0.0.0-00010101000000-000000000000 => ./build/private/bgospace/Go3p-Golang-X-Term/src/golang.org/x/term

replace golang.org/x/text v0.0.0-00010101000000-000000000000 => ./build/private/bgospace/Go3p-Golang-X-Text/src/golang.org/x/text

replace github.com/agext/levenshtein v0.0.0-00010101000000-000000000000 => ./build/private/bgospace/Go3p-Github-Agext-Levenshtein/src/github.com/agext/levenshtein

replace github.com/hashicorp/hcl/v2 v2.0.0-00010101000000-000000000000 => ./build/private/bgospace/Go3p-Github-Hashicorp-Hcl-V2/src/github.com/hashicorp/hcl/v2

replace github.com/vmihailenco/tagparser v0.0.0-00010101000000-000000000000 => ./build/private/bgospace/Go3p-Github-Vmihailenco-Tagparser/src/github.com/vmihailenco/tagparser

replace golang.org/x/net v0.0.0-00010101000000-000000000000 => ./build/private/bgospace/Go3p-Golang-X-Net/src/golang.org/x/net

replace github.com/zclconf/go-cty/cty v0.0.0-00010101000000-000000000000 => ./build/private/bgospace/Go3p-Github-Zclconf-GoCty/src/github.com/zclconf/go-cty/cty

replace github.com/zclconf/go-cty-debug/ctydebug v0.0.0-00010101000000-000000000000 => ./build/private/bgospace/Go3p-Github-Zclconf-GoCtyDebug/src/github.com/zclconf/go-cty-debug/ctydebug

replace github.com/google/go-cmp/cmp v0.0.0-00010101000000-000000000000 => ./build/private/bgospace/Go3p-Github-Google-GoCmp/src/github.com/google/go-cmp/cmp

replace github.com/apparentlymart/go-textseg/v13 v13.0.0-00010101000000-000000000000 => ./build/private/bgospace/Go3p-Github-Apparentlymart-GoTextseg-V13/src/github.com/apparentlymart/go-textseg/v13

replace golang.org/x/sys v0.0.0-00010101000000-000000000000 => ./build/private/bgospace/Go3p-Golang-X-Sys/src/golang.org/x/sys

replace github.com/mitchellh/go-wordwrap v0.0.0-00010101000000-000000000000 => ./build/private/bgospace/Go3p-Github-Mitchellh-GoWordwrap/src/github.com/mitchellh/go-wordwrap
