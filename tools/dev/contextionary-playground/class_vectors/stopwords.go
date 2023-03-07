/*                          _       _
 *__      _____  __ ___   ___  __ _| |_ ___
 *\ \ /\ / / _ \/ _` \ \ / / |/ _` | __/ _ \
 * \ V  V /  __/ (_| |\ V /| | (_| | ||  __/
 *  \_/\_/ \___|\__,_| \_/ |_|\__,_|\__\___|
 *
 * Copyright © 2016 - 2019 Weaviate. All rights reserved.
 * LICENSE: https://github.com/weaviate/weaviate/blob/master/LICENSE
 * DESIGN & CONCEPT: Bob van Luijt (@bobvanluijt)
 * CONTACT: hello@weaviate.io
 */package main

var stopWords = map[string]int{
	"a":             0,
	"able":          0,
	"about":         0,
	"above":         0,
	"abst":          0,
	"accordance":    0,
	"according":     0,
	"accordingly":   0,
	"across":        0,
	"act":           0,
	"actually":      0,
	"added":         0,
	"adj":           0,
	"affected":      0,
	"affecting":     0,
	"affects":       0,
	"after":         0,
	"afterwards":    0,
	"again":         0,
	"against":       0,
	"ah":            0,
	"all":           0,
	"almost":        0,
	"alone":         0,
	"along":         0,
	"already":       0,
	"also":          0,
	"although":      0,
	"always":        0,
	"am":            0,
	"among":         0,
	"amongst":       0,
	"an":            0,
	"and":           0,
	"announce":      0,
	"another":       0,
	"any":           0,
	"anybody":       0,
	"anyhow":        0,
	"anymore":       0,
	"anyone":        0,
	"anything":      0,
	"anyway":        0,
	"anyways":       0,
	"anywhere":      0,
	"apparently":    0,
	"approximately": 0,
	"are":           0,
	"aren":          0,
	"arent":         0,
	"arise":         0,
	"around":        0,
	"as":            0,
	"aside":         0,
	"ask":           0,
	"asking":        0,
	"at":            0,
	"auth":          0,
	"available":     0,
	"away":          0,
	"awfully":       0,
	"b":             0,
	"back":          0,
	"be":            0,
	"became":        0,
	"because":       0,
	"become":        0,
	"becomes":       0,
	"becoming":      0,
	"been":          0,
	"before":        0,
	"beforehand":    0,
	"begin":         0,
	"beginning":     0,
	"beginnings":    0,
	"begins":        0,
	"behind":        0,
	"being":         0,
	"believe":       0,
	"below":         0,
	"beside":        0,
	"besides":       0,
	"between":       0,
	"beyond":        0,
	"biol":          0,
	"both":          0,
	"brief":         0,
	"briefly":       0,
	"but":           0,
	"by":            0,
	"c":             0,
	"ca":            0,
	"came":          0,
	"can":           0,
	"cannot":        0,
	"can't":         0,
	"cause":         0,
	"causes":        0,
	"certain":       0,
	"certainly":     0,
	"co":            0,
	"com":           0,
	"come":          0,
	"comes":         0,
	"contain":       0,
	"containing":    0,
	"contains":      0,
	"could":         0,
	"couldnt":       0,
	"d":             0,
	"date":          0,
	"did":           0,
	"didn't":        0,
	"different":     0,
	"do":            0,
	"does":          0,
	"doesn't":       0,
	"doing":         0,
	"done":          0,
	"don't":         0,
	"down":          0,
	"downwards":     0,
	"due":           0,
	"during":        0,
	"e":             0,
	"each":          0,
	"ed":            0,
	"edu":           0,
	"effect":        0,
	"eg":            0,
	"eight":         0,
	"eighty":        0,
	"either":        0,
	"else":          0,
	"elsewhere":     0,
	"end":           0,
	"ending":        0,
	"enough":        0,
	"especially":    0,
	"et":            0,
	"et-al":         0,
	"etc":           0,
	"even":          0,
	"ever":          0,
	"every":         0,
	"everybody":     0,
	"everyone":      0,
	"everything":    0,
	"everywhere":    0,
	"ex":            0,
	"except":        0,
	"f":             0,
	"far":           0,
	"few":           0,
	"ff":            0,
	"fifth":         0,
	"first":         0,
	"five":          0,
	"fix":           0,
	"followed":      0,
	"following":     0,
	"follows":       0,
	"for":           0,
	"former":        0,
	"formerly":      0,
	"forth":         0,
	"found":         0,
	"four":          0,
	"from":          0,
	"further":       0,
	"furthermore":   0,
	"g":             0,
	"gave":          0,
	"get":           0,
	"gets":          0,
	"getting":       0,
	"give":          0,
	"given":         0,
	"gives":         0,
	"giving":        0,
	"go":            0,
	"goes":          0,
	"gone":          0,
	"got":           0,
	"gotten":        0,
	"h":             0,
	"had":           0,
	"happens":       0,
	"hardly":        0,
	"has":           0,
	"hasn't":        0,
	"have":          0,
	"haven't":       0,
	"having":        0,
	"he":            0,
	"hed":           0,
	"hence":         0,
	"her":           0,
	"here":          0,
	"hereafter":     0,
	"hereby":        0,
	"herein":        0,
	"heres":         0,
	"hereupon":      0,
	"hers":          0,
	"herself":       0,
	"hes":           0,
	"hi":            0,
	"hid":           0,
	"him":           0,
	"himself":       0,
	"his":           0,
	"hither":        0,
	"home":          0,
	"how":           0,
	"howbeit":       0,
	"however":       0,
	"hundred":       0,
	"i":             0,
	"id":            0,
	"ie":            0,
	"if":            0,
	"i'll":          0,
	"im":            0,
	"immediate":     0,
	"immediately":   0,
	"importance":    0,
	"important":     0,
	"in":            0,
	"inc":           0,
	"indeed":        0,
	"index":         0,
	"information":   0,
	"instead":       0,
	"into":          0,
	"invention":     0,
	"inward":        0,
	"is":            0,
	"isn't":         0,
	"it":            0,
	"itd":           0,
	"it'll":         0,
	"its":           0,
	"itself":        0,
	"i've":          0,
	"j":             0,
	"just":          0,
	"k":             0,
	"keep":          0,
	"keeps":         0,
	"kept":          0,
	"kg":            0,
	"km":            0,
	"know":          0,
	"known":         0,
	"knows":         0,
	"l":             0,
	"largely":       0,
	"last":          0,
	"lately":        0,
	"later":         0,
	"latter":        0,
	"latterly":      0,
	"least":         0,
	"less":          0,
	"lest":          0,
	"let":           0,
	"lets":          0,
	"like":          0,
	"liked":         0,
	"likely":        0,
	"line":          0,
	"little":        0,
	"'ll":           0,
	"look":          0,
	"looking":       0,
	"looks":         0,
	"ltd":           0,
	"m":             0,
	"made":          0,
	"mainly":        0,
	"make":          0,
	"makes":         0,
	"many":          0,
	"may":           0,
	"maybe":         0,
	"me":            0,
	"mean":          0,
	"means":         0,
	"meantime":      0,
	"meanwhile":     0,
	"merely":        0,
	"mg":            0,
	"might":         0,
	"million":       0,
	"miss":          0,
	"ml":            0,
	"more":          0,
	"moreover":      0,
	"most":          0,
	"mostly":        0,
	"mr":            0,
	"mrs":           0,
	"much":          0,
	"mug":           0,
	"must":          0,
	"my":            0,
	"myself":        0,
	"n":             0,
	"na":            0,
	"name":          0,
	"namely":        0,
	"nay":           0,
	"nd":            0,
	"near":          0,
	"nearly":        0,
	"necessarily":   0,
	"necessary":     0,
	"need":          0,
	"needs":         0,
	"neither":       0,
	"never":         0,
	"nevertheless":  0,
	"new":           0,
	"next":          0,
	"nine":          0,
	"ninety":        0,
	"no":            0,
	"nobody":        0,
	"non":           0,
	"none":          0,
	"nonetheless":   0,
	"noone":         0,
	"nor":           0,
	"normally":      0,
	"nos":           0,
	"not":           0,
	"noted":         0,
	"nothing":       0,
	"now":           0,
	"nowhere":       0,
	"o":             0,
	"obtain":        0,
	"obtained":      0,
	"obviously":     0,
	"of":            0,
	"off":           0,
	"often":         0,
	"oh":            0,
	"ok":            0,
	"okay":          0,
	"old":           0,
	"omitted":       0,
	"on":            0,
	"once":          0,
	"one":           0,
	"ones":          0,
	"only":          0,
	"onto":          0,
	"or":            0,
	"ord":           0,
	"other":         0,
	"others":        0,
	"otherwise":     0,
	"ought":         0,
	"our":           0,
	"ours":          0,
	"ourselves":     0,
	"out":           0,
	"outside":       0,
	"over":          0,
	"overall":       0,
	"owing":         0,
	"own":           0,
	"p":             0,
	"page":          0,
	"pages":         0,
	"part":          0,
	"particular":    0,
	"particularly":  0,
	"past":          0,
	"per":           0,
	"perhaps":       0,
	"placed":        0,
	"please":        0,
	"plus":          0,
	"poorly":        0,
	"possible":      0,
	"possibly":      0,
	"potentially":   0,
	"pp":            0,
	"predominantly": 0,
	"present":       0,
	"previously":    0,
	"primarily":     0,
	"probably":      0,
	"promptly":      0,
	"proud":         0,
	"provides":      0,
	"put":           0,
	"q":             0,
	"que":           0,
	"quickly":       0,
	"quite":         0,
	"qv":            0,
	"r":             0,
	"ran":           0,
	"rather":        0,
	"rd":            0,
	"re":            0,
	"readily":       0,
	"really":        0,
	"recent":        0,
	"recently":      0,
	"ref":           0,
	"refs":          0,
	"regarding":     0,
	"regardless":    0,
	"regards":       0,
	"related":       0,
	"relatively":    0,
	"research":      0,
	"respectively":  0,
	"resulted":      0,
	"resulting":     0,
	"results":       0,
	"right":         0,
	"run":           0,
	"s":             0,
	"said":          0,
	"same":          0,
	"saw":           0,
	"say":           0,
	"saying":        0,
	"says":          0,
	"sec":           0,
	"section":       0,
	"see":           0,
	"seeing":        0,
	"seem":          0,
	"seemed":        0,
	"seeming":       0,
	"seems":         0,
	"seen":          0,
	"self":          0,
	"selves":        0,
	"sent":          0,
	"seven":         0,
	"several":       0,
	"shall":         0,
	"she":           0,
	"shed":          0,
	"she'll":        0,
	"shes":          0,
	"should":        0,
	"shouldn't":     0,
	"show":          0,
	"showed":        0,
	"shown":         0,
	"showns":        0,
	"shows":         0,
	"significant":   0,
	"significantly": 0,
	"similar":       0,
	"similarly":     0,
	"since":         0,
	"six":           0,
	"slightly":      0,
	"so":            0,
	"some":          0,
	"somebody":      0,
	"somehow":       0,
	"someone":       0,
	"somethan":      0,
	"something":     0,
	"sometime":      0,
	"sometimes":     0,
	"somewhat":      0,
	"somewhere":     0,
	"soon":          0,
	"sorry":         0,
	"specifically":  0,
	"specified":     0,
	"specify":       0,
	"specifying":    0,
	"still":         0,
	"stop":          0,
	"strongly":      0,
	"sub":           0,
	"substantially": 0,
	"successfully":  0,
	"such":          0,
	"sufficiently":  0,
	"suggest":       0,
	"sup":           0,
	"sure":          0,
	"take":          0,
	"taken":         0,
	"taking":        0,
	"tell":          0,
	"tends":         0,
	"th":            0,
	"than":          0,
	"thank":         0,
	"thanks":        0,
	"thanx":         0,
	"that":          0,
	"that'll":       0,
	"thats":         0,
	"that've":       0,
	"the":           0,
	"their":         0,
	"theirs":        0,
	"them":          0,
	"themselves":    0,
	"then":          0,
	"thence":        0,
	"there":         0,
	"thereafter":    0,
	"thereby":       0,
	"thered":        0,
	"therefore":     0,
	"therein":       0,
	"there'll":      0,
	"thereof":       0,
	"therere":       0,
	"theres":        0,
	"thereto":       0,
	"thereupon":     0,
	"there've":      0,
	"these":         0,
	"they":          0,
	"theyd":         0,
	"they'll":       0,
	"theyre":        0,
	"they've":       0,
	"think":         0,
	"this":          0,
	"those":         0,
	"thou":          0,
	"though":        0,
	"thoughh":       0,
	"thousand":      0,
	"throug":        0,
	"through":       0,
	"throughout":    0,
	"thru":          0,
	"thus":          0,
	"til":           0,
	"tip":           0,
	"to":            0,
	"together":      0,
	"too":           0,
	"took":          0,
	"toward":        0,
	"towards":       0,
	"tried":         0,
	"tries":         0,
	"truly":         0,
	"try":           0,
	"trying":        0,
	"ts":            0,
	"twice":         0,
	"two":           0,
	"u":             0,
	"un":            0,
	"under":         0,
	"unfortunately": 0,
	"unless":        0,
	"unlike":        0,
	"unlikely":      0,
	"until":         0,
	"unto":          0,
	"up":            0,
	"upon":          0,
	"ups":           0,
	"us":            0,
	"use":           0,
	"used":          0,
	"useful":        0,
	"usefully":      0,
	"usefulness":    0,
	"uses":          0,
	"using":         0,
	"usually":       0,
	"v":             0,
	"value":         0,
	"various":       0,
	"'ve":           0,
	"very":          0,
	"via":           0,
	"viz":           0,
	"vol":           0,
	"vols":          0,
	"vs":            0,
	"w":             0,
	"want":          0,
	"wants":         0,
	"was":           0,
	"wasnt":         0,
	"way":           0,
	"we":            0,
	"wed":           0,
	"welcome":       0,
	"we'll":         0,
	"went":          0,
	"were":          0,
	"werent":        0,
	"we've":         0,
	"what":          0,
	"whatever":      0,
	"what'll":       0,
	"whats":         0,
	"when":          0,
	"whence":        0,
	"whenever":      0,
	"where":         0,
	"whereafter":    0,
	"whereas":       0,
	"whereby":       0,
	"wherein":       0,
	"wheres":        0,
	"whereupon":     0,
	"wherever":      0,
	"whether":       0,
	"which":         0,
	"while":         0,
	"whim":          0,
	"whither":       0,
	"who":           0,
	"whod":          0,
	"whoever":       0,
	"whole":         0,
	"who'll":        0,
	"whom":          0,
	"whomever":      0,
	"whos":          0,
	"whose":         0,
	"why":           0,
	"widely":        0,
	"willing":       0,
	"wish":          0,
	"with":          0,
	"within":        0,
	"without":       0,
	"wont":          0,
	"words":         0,
	"world":         0,
	"would":         0,
	"wouldnt":       0,
	"www":           0,
	"x":             0,
	"y":             0,
	"yes":           0,
	"yet":           0,
	"you":           0,
	"youd":          0,
	"you'll":        0,
	"your":          0,
	"youre":         0,
	"yours":         0,
	"yourself":      0,
	"yourselves":    0,
	"you've":        0,
	"z":             0,
	"zero":          0,
}

func isStopWord(word string) bool {
	if _, ok := stopWords[word]; ok {
		return true
	}

	return false
}
