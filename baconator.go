package baconator

import (
	"bufio"
	"compress/bzip2"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/willabides/baconator/internal/graph"
)

type nodeType int8

const (
	castNode nodeType = iota + 1
	movieNode
)

type movie struct {
	Year  int      `json:"year"`
	Title string   `json:"title"`
	Cast  []string `json:"cast"`
}

type nodeInfo struct {
	Node graph.NodeIdx
	Type nodeType
	Name string
}

// Baconator is kind of a big deal around here
type Baconator struct {
	CastNodes  map[string]graph.NodeIdx
	MovieNodes map[string]graph.NodeIdx
	NodeInfo   []nodeInfo
	Graph      *graph.Graph
	Movies     map[string]*movie
}

// LoadFromDatafile loads b with data in filename
func (b *Baconator) LoadFromDatafile(filename string) error {
	err := downloadDataIfNeeded(filename)
	if err != nil {
		return fmt.Errorf("error downloading data file: %v", err)
	}
	movies, err := loadMovies(filename)
	if err != nil {
		return err
	}
	bac := buildBaconator(movies)
	*b = *bac
	return nil
}

type stringNeighbors map[string]map[string]bool

func (n stringNeighbors) sortedKeys() []string {
	keys := make([]string, 0, len(n))
	for s := range n {
		keys = append(keys, s)
	}
	sort.Strings(keys)
	return keys
}

func (n stringNeighbors) sortedValues(key string) []string {
	vals := make([]string, 0, len(n[key]))
	for val := range n[key] {
		vals = append(vals, val)
	}
	sort.Strings(vals)
	return vals
}

func downloadDataIfNeeded(filename string) (err error) {
	err = os.MkdirAll(filepath.Dir(filename), 0o700)
	if err != nil {
		return err
	}
	_, err = os.Stat(filename)
	if !os.IsNotExist(err) {
		return err
	}
	u := "https://oracleofbacon.org/data.txt.bz2"
	res, err := http.Get(u)
	if err != nil {
		return err
	}
	if res.StatusCode != 200 {
		return fmt.Errorf("unexpected http status: %d", res.StatusCode)
	}
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer func() {
		cerr := file.Close()
		if err == nil {
			err = cerr
		}
	}()
	_, err = io.Copy(file, res.Body)
	if err != nil {
		return err
	}
	return res.Body.Close()
}

func loadMovies(filename string) (map[string]*movie, error) {
	result := map[string]*movie{}
	file, err := os.Open(filename) //nolint:gosec // not user supplied
	if err != nil {
		return nil, err
	}
	scanner := bufio.NewScanner(bzip2.NewReader(file))
	for scanner.Scan() {
		var film movie
		err = json.Unmarshal(scanner.Bytes(), &film)
		if err != nil {
			return nil, err
		}
		if result[film.Title] != nil {
			return nil, fmt.Errorf("duplicate title: %q", film.Title)
		}
		result[film.Title] = &film
	}
	return result, nil
}

func parseCastName(nm string) string {
	nm = strings.TrimPrefix(nm, "[[")
	nm = strings.TrimSuffix(nm, "]]")
	parts := strings.Split(nm, "|")
	return parts[len(parts)-1]
}

func buildBaconator(movies map[string]*movie) *Baconator {
	movieCast, castMovies := buildNeighbors(movies)

	b := Baconator{
		CastNodes:  make(map[string]graph.NodeIdx, len(movieCast)),
		MovieNodes: make(map[string]graph.NodeIdx, len(movieCast)),
		NodeInfo:   make([]nodeInfo, 0, len(movieCast)+len(castMovies)),
		Movies:     movies,
	}

	for _, title := range movieCast.sortedKeys() {
		if _, ok := b.MovieNodes[title]; !ok {
			node := graph.NodeIdx(len(b.NodeInfo))
			b.NodeInfo = append(b.NodeInfo, nodeInfo{
				Node: node,
				Type: movieNode,
				Name: title,
			})
			b.MovieNodes[title] = node
		}
		cast := movies[title].Cast
		for i, s := range cast {
			cast[i] = parseCastName(s)
		}
		sort.Strings(cast)
		for _, castMember := range cast {
			if _, ok := b.CastNodes[castMember]; ok {
				continue
			}
			node := graph.NodeIdx(len(b.NodeInfo))
			b.NodeInfo = append(b.NodeInfo, nodeInfo{
				Node: node,
				Type: castNode,
				Name: castMember,
			})
			b.CastNodes[castMember] = node
		}
	}
	b.Graph = b.buildGraph(movieCast, castMovies)
	return &b
}

func (b *Baconator) buildGraph(movieCast, castMovies stringNeighbors) *graph.Graph {
	neighborhood := make([][]graph.NodeIdx, len(b.NodeInfo))
	for n := graph.NodeIdx(0); int(n) < len(b.NodeInfo); n++ {
		info := b.NodeInfo[n]
		switch info.Type {
		case movieNode:
			cast := movieCast.sortedValues(info.Name)
			neighborhood[n] = make([]graph.NodeIdx, 0, len(cast))
			for _, castMember := range cast {
				nn, ok := b.CastNodes[castMember]
				if !ok {
					panic("node doesn't exist")
				}
				neighborhood[n] = append(neighborhood[n], nn)
			}
			sortNodes(neighborhood[n])
		case castNode:
			movies := castMovies.sortedValues(info.Name)
			neighborhood[n] = make([]graph.NodeIdx, 0, len(movies))
			for _, film := range movies {
				nn, ok := b.MovieNodes[film]
				if !ok {
					panic("node doesn't exist")
				}
				neighborhood[n] = append(neighborhood[n], nn)
			}
			sortNodes(neighborhood[n])
		default:
			panic("unexpected node type")
		}
	}
	return graph.New(neighborhood)
}

func sortNodes(nodes []graph.NodeIdx) {
	sort.Slice(nodes, func(i, j int) bool { return nodes[i] < nodes[j] })
}

func buildNeighbors(movies map[string]*movie) (movieCast, castMovies stringNeighbors) {
	movieCast = stringNeighbors{}
	castMovies = stringNeighbors{}
	for _, m := range movies {
		if len(m.Cast) == 0 {
			continue
		}
		movieCast[m.Title] = make(map[string]bool, len(m.Cast))
		for _, c := range m.Cast {
			c = parseCastName(c)
			movieCast[m.Title][c] = true
			if castMovies[c] == nil {
				castMovies[c] = map[string]bool{}
			}
			castMovies[c][m.Title] = true
		}
	}
	return movieCast, castMovies
}

type linksResult struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

func (b *Baconator) links(src, dest string) ([]linksResult, error) {
	srcNode, ok := b.CastNodes[src]
	if !ok {
		return nil, fmt.Errorf("unknown cast member: %q", src)
	}
	destNode, ok := b.CastNodes[dest]
	if !ok {
		return nil, fmt.Errorf("unknown cast member: %q", dest)
	}
	var path []graph.NodeIdx
	var pri graph.PriorityFunc = func(node graph.NodeIdx) int64 {
		info := b.NodeInfo[node]
		if info.Type != movieNode {
			return 0
		}
		year := 10000
		film := b.Movies[info.Name]
		if film != nil && film.Year > 0 {
			year = film.Year
		}
		return int64(year * -1)
	}
	b.Graph.FindPath(&path, 99, srcNode, destNode, pri)
	res := make([]linksResult, len(path))
	for i, node := range path {
		info := b.NodeInfo[node]
		res[i].Name = info.Name
		switch info.Type {
		case movieNode:
			res[i].Type = "movie"
		case castNode:
			res[i].Type = "cast"
		default:
			res[i].Type = "unknown"
		}
	}
	return res, nil
}

type centerResult struct {
	Distance    map[int]int `json:"count_by_distance"`
	Total       int         `json:"total_linkable"`
	AvgDistance float64     `json:"average_distance"`
}

func (b *Baconator) center(center graph.NodeIdx) *centerResult {
	result := centerResult{
		Distance: map[int]int{},
	}
	levels := b.Graph.FindLevels(center)
	var maxLevel int
	if b.NodeInfo[center].Type != castNode {
		return nil
	}
	for i, level := range levels {
		if b.NodeInfo[i].Type != castNode {
			continue
		}
		result.Total++
		if level > maxLevel {
			maxLevel = level
		}
	}
	var tot float64
	for i, level := range levels {
		if level == 0 {
			continue
		}
		if b.NodeInfo[i].Type != castNode {
			continue
		}
		tot += float64(level / 2)
		result.Distance[level/2]++
	}
	result.AvgDistance = tot / float64(result.Total)
	return &result
}
