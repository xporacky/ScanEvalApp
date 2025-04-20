package main_test

import (
	"ScanEvalApp/internal/database/repository"
	"ScanEvalApp/internal/database/seed"
	"ScanEvalApp/internal/logging"
	"ScanEvalApp/internal/scanprocessing"

	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var expectedResults = map[int]string{
	2:  "acdcebcdbcdcbecbbcbdacbcdedecdcabcbcdedc",
	3:  "acbbexcbdbaabcccbedeabbccceadbdcebcdbced",
	32: "bcadabdcabecbcaedabdbcdabdeabdbcdaeadcbd",
	27: "acxebbbcbbcdeaabccaebecbdabcdcaabbcbcdcb",
	22: "abaceabcxadbeaeabcdecccccbeeaabcadaedabd",
	35: "dcbaedcbaedcbaedcbaebacecabccecabdcbdaec",
	13: "cdbdcbbcdddccbcabdbadbadedbcedbcbcdxbcea",
	24: "abdddeebacbaebdabcedabeddcbaedabecdabccc",
	47: "cedbabcxddbabdeabcdaabcxeccbxebbbdccccba",
	14: "abcededbacabcdeddddddbceaabcedaedbccbdae",
	49: "acbdcbcdccabdcbeccbeecdbcadcbdedddccceba",
	18: "abceedcebaabedcabcdebacddedcbaabcdeaaaaa",
	15: "cdbeacbaedacedbabcdeabecdabcedbbbbbcdeab",
	44: "dcdbeabcccddbcbadcbaaccbdedacdaedcdcbbxa",
	45: "aedbcacebdabdecbcaeaaeaeabdbdbcacacdcdcd",
	43: "abccbcdcdeebcecdbebaabcdxdccaecbcebbddee",
	41: "abcdaedecabcdeabacdeeeeeeabcdebdacebdeac",
	8:  "addcaabecxbcddcdbacaabbbcbbccdabebeedcec",
	29: "abcdebbbbbcccccdddddabcdedacebabdecbacde",
	21: "abcebcdccaabccebbcdecacddebaxabcccdedcba",
	12: "edcdddcbbaedccbbedddabbcdcccbbbcddcecdba",
	38: "dabcedecababceeabcdexxxxxxxxxxxxxxxxxxxx",
	7:  "abdceabceabcdeabacdeabceeabcdeabcdeaaaaa",
	4:  "edcbxcebdabcdbxadcdaabdcdcccccdacbebcbde",
	20: "abcedabcddedcabacdeabaedcabcdceabdeacbed",
	34: "edcddcbdcbeddbcacbbebccaedecbcbcbdcabcdd",
	46: "abecedccbacbbddbddcaeddcbeedcxacbxeabced",
	26: "aabcecdbaaeeeecddabeabccddceabacedabcdea",
	39: "ebcaebcbaecbbdccabdcabxbdbdcbxcdbcaabxce",
	9:  "cbdbeabecxabdeabcdcexxxxxxxxxxxxxxxxxxxx",
	16: "abcdeabcdeabcdedceabaabcceecdbcabecbceab",
	6:  "edcbabcbcecbdaeecbababcceabcbabbbdeccbae",
	42: "abbbbbbbacdedbabcdcexxxxxxxxxxxxxxxxxxxx",
	5:  "acdcecbadbdcbccabbeeaaccecabceabaadaddce",
	36: "abcaedcabdcbdaedbacdeeeeedcxabcdeabcaade",
	25: "bcaedabcdeabcdeabbccdddddbbbbbcdaceabced",
	23: "abcdceedcdabdeabcdcexabcecbcabacdeabecbc",
	17: "acbdeabcedabcedabdceabdcabcdaaxcaedbadca",
	40: "abababcdbcdbcdbcdeababceacbdaeabdecabdca",
	19: "abcdaabcaabcabeddeddababacdcdcxdededcbac",
	31: "bcaedbacdedebacbaacecdbeabcedabcdeabdcae",
	1:  "abcdeabcdeedcbaabcdeabcdeedcbaabcdeedcba",
	37: "ebadecbdabcdbaebcdccabcbdabcdabcdeabcdea",
	28: "abdccedcbcbbbbbdededaedbcebacedbceabcdea",
	10: "aadcebecdbabeabcxbecaadecbabedcdbadcdcba",
	11: "cdaebdcacxdbceeabbccbbdddaaeedcbabdedcbb",
	50: "bcaxedcbceaaaaabcdddabcddcbbcabdcdeeaabb",
	30: "dcdcedabcdbacceaacaaedcbabcdeeddccbbaaee",
	48: "eddcdxaabcbabacadaedbcaedcdebaabacaeddcb",
	33: "baacecbdxbcbdaaeecdabbbdacedabcbaecabcce",
}

var expectedAnswer_923 = map[int]string{
	392411: "bccdeeddccdddcxcxceddcxebxcceddbeababddc",
	913602: "aaaaabbbbbcccccdddddbcbcdbcdbcbababxxxxx",
	985335: "abdcxbcdcdbxxaedxdcbabccbcbcdecbabcccbbc",
	257457: "abdcaxdeccbcxaebbxecabcccccccbdcbadxaccd",
	133168: "abcdeacdedxbdcbcdcdcbcdedcbacedcbabdedcb",
	526606: "abcbacbcdcbcdbcbdxcxbbbbbxbbbxbbbcdxcxbc",
	411098: "bdbcdbaaaabbbbbcccccdacdbcxdxdcxdcbdbxce",
	631788: "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
	433835: "bedcdcdeccbeadcxxxxxaaabeabbbcbcdecaaaaa",
	300532: "abcccbbcccbbbbbdddddeeeeeeeeeedddddaaaaa",
	801650: "abbbbddddddcxabadcdbaaaaabbbbbcccccbbbbb",
	783424: "abccccccccbbbbbacbcdabbbbccccceeeeeaaaaa",
	990337: "eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee",
	753491: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	648037: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	188319: "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
	988443: "xxxxxxxxxxxxxxcxxxxxabcdexxxxxxxxxxxxxxx",
	624245: "abcdeeeeeeeeeeeeeeeexdxxxexdeeeeeeeadxed",
	774776: "xdaaaaaaaaaaabbbcdeeabcddeeeeeeeeeeeeeee",
	236273: "aaaaaaaaaaaaaxaaaaaaaaaaaaaaaaaaaaaaaaaa",
	537282: "cccccccccccccccccccccccccccccccccccccccc",
	639863: "cccccccccccccccccccccccccccccccccccccccc",
	227633: "cccccccccccccccccccccccccccccccccccccccc",
	872413: "cccccccccccccccccccccccccccccccccccccccc",
	212971: "abcxcdxcxcdeddcecdedaxcdebbddedddedxcccd",
	507113: "xcccxxbdddddddccccxbcccccccccccccccccccc",
	152342: "dddddddddddddddddddddddddddddddddddddddd",
	870991: "abbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
	783456: "ccccccccccccccccccccdddcccbbbbbbbaxxxdec",
	590787: "bbbbbbbbbbbbbbbbbbbbcccccccccccccccccccc",
	705932: "abababcbcbcdcdcdededabababcbcbcdcdcdeded",
	363580: "abcdeabcdeabcdeabcdeabcdeabcdeabcdeabcde",
	168030: "bdcaabccccdcccccccccedcbaabcdedcbaaabxde",
	646042: "bdecacdecccbdccxxbccaaaaaaaaaaaaaaaaaaaa",
	375942: "abcdeabcdeabcdeabcdeabcdeabcdeabcdeabcde",
	658013: "bcdecbcdedbcdddbbbbbbbbbbcbbbbbbbbbbbbbb",
	798399: "abcdeabcdeabcdeabcdeabcdeabcdeabcdeabcde",
	984996: "abcdebcdeeabcdeabcdeabcdeabcdeabcdeabcde",
	484470: "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxaaaxxxxxxxx",
	731579: "abcdeabcdeabcdeabcdeabcdeabcdeabcdeabcde",
	235850: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	404065: "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
	760018: "ccccccccccccccccccccxxxxxbccdddddddddddd",
	422417: "abcdeabcdeabcdeabcdeabcdeabcdeabcdeabcde",
	636975: "bbbbbbbbbbbbbbbbbbbbxxxxxxxxxxxxxxxxxxxx",
	567742: "abcdeabcdeabcdeabcdeabcdeabcdeabcdeabcde",
	123800: "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
	385676: "bbbbbbbbbbbbbbbbbbbbbbbbbcccccbbbbbaaaaa",
	761336: "abcdeabcdeabcdeabcdeabcdeabcdeabcdeabcde",
	110198: "dabcdcccccdeexdabcdeabcdeabcdeedcbbabxed",
}

var expectedAnswer_623 = map[int]string{
	507113: "babababababababababababababababababababx",
	227633: "xxxxccccccccccccccccxxxxxxxxxxxxxxxxxccc",
	152342: "cccccccccccccccccccccccccccccccccccccccc",
	870991: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
	590787: "dddddddddddddddddddddddddddddddddddddddd",
	646042: "eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee",
	984996: "eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee",
	404065: "dddddddddddddddddddddddddddddddddddddddd",
	567742: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	385676: "abbcxabcdexxxxxabcdeabcdeabcdeabxxeabcde",
	774776: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	375942: "aaaxxxxbbbbbbbbbbbbbxxxxxxxbxxxxxxbxxbxx",
	433835: "dddddddddddddddddddddddddddddddddddddddd",
	411098: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	526606: "bbbbbbbbbbbbbbbbbbbbccccccccccccccbbbbbb",
	257457: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	783456: "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
	801650: "ebccbcexcdxecbdbbbcdaaaaaadcbccdaabccccc",
	753491: "cebdaababcaaaaabcccdecbcdabcdcbbbbbccddd",
	110198: "aaaaaccccceeeeedbdbdaabcdcbbcdbbbbbccccc",
	705932: "bbbbbbbbbbbbbbbbbbbbxxxxxxxxxxxxxxxxxxxx",
	537282: "abcdeeeeeedcbaeedcbaabcdedcbabcdeedcbabc",
	484470: "eeeeeeeeeeeeeeeeeeeeabcdeedcbaabcdeedcba",
	422417: "abcdeeedcbabcdedcbaxxbcxeeeeeeeeeeeeeeee",
	761336: "eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee",
	392411: "xxxxxxxxxxxxxxxxxxxxdedeeeeeeeeeeeeeeeee",
	212971: "abcdeeeddedeeeexxeeexxxxxxxxxxxxxxxxxxxx",
	783424: "abcdeedcbabcdedcbabcbbbbbdccccxddddxaaaa",
	133168: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
	648037: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	872413: "abcaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	639863: "abcddaaaaabbbbbdddddbabcdabcdeccaaabbbbb",
	363580: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	985335: "ceececdxcdcadddcddcbcebedcdeedcbabddcdee",
	168030: "eeeeecccccbbbbbxxxxxcccccbbbbbaaaaabcccc",
	235850: "abcdeaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	798399: "aaaaabbbbbcccccdddddaaaaaeeeeedddddbbbbb",
	731579: "bbbbbbbaaabbababbbbbeeeeeeeeeedddddddddd",
	760018: "aaaaabbbbbcccccdddddbbbbbcccccdddddccccc",
	636975: "abbbcbbcbcbaaaaccdccaaaaabbbbbcdaaeabcbc",
	658013: "cccccccccccccccccccccccccccccccccccccccc",
	631788: "eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee",
	990337: "abcdeabcdeabcdeabcdeabcdeabcdeabcdeabcde",
	236273: "abcdeabcdeabcdeabcdeabcdeabcdeabcdeabcde",
	300532: "abcdedcdedcbabcdedcbabcdedcbabcdedcbabcd",
	988443: "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
	123800: "aaaaabbbbbccccccccccaaaaacccccdddddbcdcb",
	624245: "aaaaabcdeababcdabcdeaabdcabcdcabcdcabbbb",
	913602: "abcdeaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	188319: "bbbbbbbbbbbbbbbbbbbbaaaaaaaaaaaaaaaaaaaa",
}

// test 5 (5 marec)
var expectedAnswer_130 = map[int]string{
	507113: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	227633: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	152342: "cccccccccccccccccccccccccccccccccccccccc",
	870991: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
	590787: "dddddddddddddddddddddddddddddddddddddddd",
	646042: "cccccccccccccccccccccccccccccccccccccccc",
	984996: "ceabcedecdeexedeeecdbcdededcbcdedcbabcde",
	404065: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	567742: "aaaaaaaaaaaaaaaaaaaaabababababababababab",
	385676: "babbbcccccbbbbbcccccaaaaabbbbbcccccccccc",
	774776: "aaaaaaaaaabbbbbcccccdddddeeeeeaaaaabbbbb",
	375942: "ababababababababbbbbcdecdecdecdecdecdedd",
	433835: "aaaaabbbbbaaaaabbbbbdddddeeeeeaaaaabbbbb",
	411098: "bbbbbcccccdddddeeeeeeeeeedddddcccccbbbbb",
	526606: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
	257457: "abababababababababababababababababababab",
	783456: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	801650: "cddcdcddcccccccddddddddddcccccaaaaaaaaaa",
	753491: "bcbcbcbcbcbcbcbcbcbcebecddeedddeeeeddddd",
	110198: "abcdedcbabcdedcbabcdabcdedcbabcdedcbabcd",
	705932: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	537282: "eeeeedddddcccccbbbbbeeeeeeeeeeeeeeeeeeee", // opilec vyplnal (zle kriziky)
	484470: "eeeddcccccaaabbbbbabdddddcccccbbbbbaaaaa",
	422417: "aabbaabbaabbaabbaabbdddddddddddddddeeeee",
	761336: "aaaaabbbbbcccccdddddeeeeeeeeeeeeeeeccccc", // opilec
	392411: "cbbbbxddddeeeeeaaaaacccccdddddaaaaaccbcc",
	212971: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	783424: "ededededededededededeeeeeeeeeeeeeeeeeeee",
	133168: "abcdeabcdeabcdeabcdeedcbaedcbaedcbaedcba",
	648037: "dddddcccccbbbbbaaaaadddddeeeeeaaaaaeeeee",
	872413: "cccccccccccccccccccccccccccccccccccccccc",
	639863: "abcdedcbabcdedcbabcdabcdeedcbabcdeeeeeee",
	363580: "abcdedcbabcdedcbabcdabcdedcbabcdeeedcbab",
	985335: "eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee",
	168030: "aaaaaaaaaaaaaaaaaaaaaaaaabaaaaaaaaaaaaaa", // tu bola dopisana 21 v papieri
	235850: "dcdcdcdcdcdcdcdcdcdcbcbcbcbcbcbcbcbcbcbb", // predposledne mozno x 34slide
	798399: "edcbabcdedcbabcdedcbabcdedcbabcdedcbabcx",
	731579: "bbcccdddddeeeeeaaaaabbbbbbbbbbcccccddddd",
	760018: "aaaaabbbbbcccccdddddcccccdddddeeeeeccccc",
	636975: "cccccbbbbbaaaaadddddeeeeedddddcccccbbbbb",
	658013: "aaaaabbbbbbbbbbcccccdededddccdccccddcccc",
	631788: "aaaaaaaaaaaaaaaaaaaacccccccccccccccccccc",
	990337: "ababababababababababcbcbcbcbcbcbcbcbcbcb",
	236273: "cbadecccccbbbbbaaaaaeeeeedddddcccccbbbbb",
	300532: "ababababababababababdededededededededede",
	988443: "dededededecbcbbaaaaacccccbbbbbaaaaaeeeee",
	123800: "bbbbbbbbbbbbbbbbbbbbcccccccccccccccccccc",
	624245: "abcdeedcbaabcdeedcbadededededededededede",
	913602: "dddddddddddddddddddddddddddddddddddddddd",
	188319: "aaaaabbbbbaaaaabbbbbaaaaabbbbbaaaaaccccc",
}

// test 4 (5 marec)
var expectedAnswer_190 = map[int]string{
	507113: "dddddddddddddddddddddddddddddddddddddddd",
	227633: "eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee",
	152342: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	870991: "dddddddddddddddddddddddddddddddddddddddd",
	590787: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
	646042: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	984996: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	404065: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	567742: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	385676: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	774776: "aaaaaaaabbbbbbbbccccccccddddddddeeeeeeee",
	375942: "cccccccccccccccccccccccccccccccccccccccc",
	433835: "aaaabbbbaaaabbbbaaaabbbbaaaabbbbaaaabbbb",
	411098: "edcbaedcbaedcbaedcbaedcbaedcbaedcbaedcba",
	526606: "aaaabbbbccccddddeeeeaaaabbbbccccddddeeee",
	257457: "eeeeeddddccccbbbbaaaeeeeddddccccbbbbaaaa",
	783456: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	801650: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	753491: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
	110198: "abababababababababababababababababababab",
	705932: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	537282: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	484470: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	422417: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	761336: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	392411: "abababcbcbcdcdcdedededededcdcdcbcbcbabab",
	212971: "aabbccddeeddccbbaaaaabcdeedcbaabcdeecbac",
	783424: "cccccccccccccccccccccccccccccccccccccccc",
	133168: "dddddddddddddddddddddddddddddddddddddddd",
	648037: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	872413: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	639863: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	363580: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	985335: "aaaaabbbbbcccccdddddaaaaabbbbbcccccddddd",
	168030: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	235850: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	798399: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	731579: "dddddddddddddddddddddddddddddddddddddddd",
	760018: "cccccccccccccccccccccccccccccccccccccccc",
	636975: "cccccccccccccccccccccccccccccccccccccccc",
	658013: "eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee",
	631788: "eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee",
	990337: "aaaaaaaabbbbbbbbccccccccddddddddeeeeeeee",
	236273: "cccccccccccccccccccccccccccccccccccccccc",
	300532: "aaaaaaaaaabbbbbbbbbbccccccccccdddddddddd",
	988443: "aaaaaaaaeeeeeeeeccccccccbbbbbbbbdddddddd",
	123800: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	624245: "aaaaaaaabbbbbbbbccccccccbbbbbbbbaaaaaaaa",
	913602: "abcdeabcdeabcdeabcdeabcdeabcdeabcdeabcde",
	188319: "eeeeeeeeddddddddccccccccbbbbbbbbaaaaaaaa",
}

var expectedAnswer_9april = map[int]string{
	507113: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	227633: "eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee",
	152342: "eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee",
	870991: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	590787: "eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee",
	646042: "eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee",
	984996: "dddddddddddddddddddddddddddddddddddddddd",
	404065: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	567742: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	385676: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	774776: "dddddddddddddddddddddddddddddddddddddddd",
	375942: "dddddddddddddddddddddddddddddddddddddddd",
	433835: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	411098: "cccccccccccccccccccccccccccccccccccccccc",
	526606: "dddddddddddddddddddddddddddddddddddddddd",
	257457: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	783456: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	801650: "eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee",
	753491: "cccccccccccccccccccccccccccccccccccccccc",
	110198: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
	705932: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	537282: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	484470: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	422417: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	761336: "eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee",
	392411: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	212971: "aaaabbbbccccddddeeeeaaaabbbbccccddddeeee",
	783424: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	133168: "eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee",
	648037: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	872413: "eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeedexe",
	639863: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	363580: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	985335: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
	168030: "eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee",
	235850: "cccccccccccccccccccccccccccccccccccccccc",
	798399: "eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee",
	731579: "eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee",
	760018: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
	636975: "eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee",
	658013: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	631788: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
	990337: "eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee",
	236273: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	300532: "eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee",
	988443: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	123800: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	624245: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	913602: "cccccccccccccccccccccccccccccccccccccccc",
	188319: "eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee",
}

func setupTestDB() (*gorm.DB, error) {
	testDBPath := "./internal/database/scan-eval-test-db.db"
	db, err := gorm.Open(sqlite.Open(testDBPath), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return db, nil
}

func getTestFilePath(relativePath string) string {
	basePath, _ := os.Getwd()
	return filepath.Join(basePath, "./assets/tmp", relativePath)
}

func TestAnswerRecognition(t *testing.T) {
	pdfPath := getTestFilePath("scan-pdfs/9_April/600.pdf")
	fmt.Printf(pdfPath)
	errorLogger := logging.GetErrorLogger()
	db, err := setupTestDB()
	if err != nil {
		errorLogger.Error("Nepodarilo sa pripojiť k databáze", slog.Group("CRITICAL", slog.String("error", err.Error())))
		t.Fatalf("Nepodarilo sa pripojiť k databáze: %v", err)
	}
	students, err := repository.GetAllStudents(db)
	if err != nil {
		return
	}
	for _, student := range students {
		student.Answers = seed.GenerateAnswers(40)
		repository.UpdateStudent(db, &student)
	}
	exam, err := repository.GetExam(db, 1) // testID = 1
	if err != nil {
		t.Fatalf("Nepodarilo sa načítať skúšku: %v", err)
	}
	startTime := time.Now()
	scanprocessing.ProcessPDF(pdfPath, exam, db, nil)
	duration := time.Since(startTime)

	totalQuestions := 0
	totalCorrect := 0
	totalMissing := 0
	totalUnrecognized := 0

	for studentID, expectedAnswers := range expectedAnswer_9april {
		student, err := repository.GetStudentByRegistrationNumber(db, uint(studentID), 1)
		if err != nil {
			t.Errorf("Študent %d nebol nájdený: %v\n", studentID, err)
			totalMissing += len(expectedAnswers)
			continue
		}
		fmt.Printf("-----------------------\n")
		recognizedAnswers := student.Answers
		if len(recognizedAnswers) == 0 {
			t.Errorf("Študent %d: chýbajúce odpovede\n", studentID)
			totalMissing += len(expectedAnswers)
			continue
		}

		correctCount := 0
		missingCount := 0
		unrecognized := 0
		totalQuestions += len(expectedAnswers)

		for i := 0; i < len(expectedAnswers); i++ {
			if i >= len(recognizedAnswers) {
				t.Errorf("Študent %d, otázka %d: chýbajúca odpoveď\n", studentID, i+1)
				missingCount++
				continue
			}

			if recognizedAnswers[i] == expectedAnswers[i] {
				correctCount++
			} else if recognizedAnswers[i] == '0' {
				totalUnrecognized++
				unrecognized++
			}
		}

		totalCorrect += correctCount
		totalMissing += missingCount
		fmt.Printf("Študent %d: správne %d/40, chýbajúce %d, nezachytené %d\n", studentID, correctCount, missingCount, unrecognized)

	}

	successRate := float64(totalCorrect) / float64(totalQuestions) * 100
	fmt.Printf("Celková úspešnosť OCR: %.2f%% (%d/%d správnych odpovedí, %d chýbajúcich, %d nezachytených)\n", successRate, totalCorrect, totalQuestions, totalMissing, totalUnrecognized)
	fmt.Printf("Čas vyhodnotenia: %.2fs\n", duration.Seconds())
}

func TestStudentAnswersExistence(t *testing.T) {
	errorLogger := logging.GetErrorLogger()
	db, err := setupTestDB()
	if err != nil {
		errorLogger.Error("Nepodarilo sa pripojiť k databáze", slog.Group("CRITICAL", slog.String("error", err.Error())))
		t.Fatalf("Nepodarilo sa pripojiť k databáze: %v", err)
	}
	totalStudents := len(expectedResults)
	recognizedCount := 0
	for studentID := range expectedResults {
		student, err := repository.GetStudentById(db, uint(studentID), 1) // testID = 1
		if err != nil {
			t.Errorf("Študent %d nebol nájdený: %v", studentID, err)
			continue
		}
		recognizedCount++

		answers := student.Answers

		count := strings.Count(answers, "0")
		recognizedAnswers := len(answers) - count
		fmt.Printf("Študent %d: rozpoznané odpovede %d/%d\n", studentID, recognizedAnswers, len(answers))

	}
	fmt.Printf("Celkový počet študentov: %d, rozpoznaných študentov: %d\n", totalStudents, recognizedCount)

}

func TestMissingPages(t *testing.T) {
	errorLogger := logging.GetErrorLogger()
	db, err := setupTestDB()
	if err != nil {
		errorLogger.Error("Nepodarilo sa pripojiť k databáze", slog.Group("CRITICAL", slog.String("error", err.Error())))
		t.Fatalf("Nepodarilo sa pripojiť k databáze: %v", err)
	}

	missingPages := 0

	for studentID := range expectedResults {
		student, err := repository.GetStudentById(db, uint(studentID), 1) // testID = 1
		if err != nil {
			t.Errorf("Študent %d nebol nájdený: %v", studentID, err)
			missingPages++
			missingPages++
			continue
		}

		recognizedAnswers := student.Answers

		zeroCount := strings.Count(recognizedAnswers, "0")

		if zeroCount == 40 {
			missingPages += 2
		}
		if zeroCount == 20 {
			missingPages++
		}
	}

	fmt.Printf("Celkový počet chýbajúcich strán: %d\n", missingPages)
}
