package cli

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/gosimple/slug"
	"github.com/kr/pretty"
	"github.com/nanoteck137/packer/utils"
	"github.com/spf13/cobra"
)

var packCmd = &cobra.Command{
	Use: "pack",
}

type Info struct {
	Name   string `json:"name"`
	Series string `json:"series"`

	IsManga        bool `json:"isManga"`
	PreferVertical bool `json:"preferVertical"`

	Cover string   `json:"cover"`
	Pages []string `json:"pages"`
}

type MangaInfoChapter struct {
	Index int      `json:"index"`
	Name  string   `json:"name"`
	Pages []string `json:"pages"`
}

type MangaInfo struct {
	Title    string             `json:"title"`
	Chapters []MangaInfoChapter `json:"chapters"`
}

func ReadMangaInfo(p string) (MangaInfo, error) {
	d, err := os.ReadFile(p)
	if err != nil {
		return MangaInfo{}, err
	}

	var res MangaInfo
	err = json.Unmarshal(d, &res)
	if err != nil {
		return MangaInfo{}, err
	}

	return res, nil
}

var packOldManga = &cobra.Command{
	Use: "old-manga <BASE> <OUTPUT>",
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		base := args[0]
		out := args[1]

		// NOTE(patrik):
		//   <NAME>.swe
		//     info.json
		//     *PAGES*.jpg|png
		//     cover.png 80x112

		mangaInfo, err := ReadMangaInfo(path.Join(base, "manga.json"))
		if err != nil {
			log.Fatal(err)
		}

		pretty.Println(mangaInfo)

		for _, c := range mangaInfo.Chapters {
			func() {
				p := path.Join(base, "chapters", strconv.Itoa(c.Index))
				fmt.Printf("p: %v\n", p)

				name := strings.TrimSpace(c.Name)

				// TODO(patrik): Validate the whole manga info
				if name == "" {
					log.Fatal("Name can't be empty")
				}

				fname := slug.Make(name)

				f, err := os.OpenFile(path.Join(out, fname+".sw"), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
				if err != nil {
					log.Fatal(err)
				}
				defer f.Close()

				w := zip.NewWriter(f)
				defer w.Close()

				var pages []string
				for i, page := range c.Pages {
					p := path.Join(p, page)

					fmt.Printf("p: %v\n", p)

					// TODO(patrik): Check ext for jpeg png
					ext := path.Ext(p)
					n := strconv.Itoa(i) + ext
					pages = append(pages, n)

					h := &zip.FileHeader{
						Name: n,
					}

					w, err := w.CreateHeader(h)
					if err != nil {
						log.Fatal(err)
					}

					src, err := os.Open(p)
					if err != nil {
						log.Fatal(err)
					}

					_, err = io.Copy(w, src)
					if err != nil {
						log.Fatal(err)
					}
				}

				dname, err := os.MkdirTemp("", "packer")
				if err != nil {
					log.Fatal(err)
				}
				defer os.RemoveAll(dname)

				coverDst := path.Join(dname, "cover.png")
				err = utils.CreateResizedImage(path.Join(p, c.Pages[0]), coverDst, 80, 112)
				if err != nil {
					log.Fatal(err)
				}

				cw, err := w.Create("cover.png")
				if err != nil {
					log.Fatal(err)
				}

				src, err := os.Open(coverDst)
				if err != nil {
					log.Fatal(err)
				}

				_, err = io.Copy(cw, src)
				if err != nil {
					log.Fatal(err)
				}

				info := Info{
					Name:           c.Name,
					Series:         mangaInfo.Title,
					IsManga:        true,
					PreferVertical: false,
					Cover:          "cover.png",
					Pages:          pages,
				}

				iw, err := w.Create("info.json")
				if err != nil {
					log.Fatal(err)
				}

				e := json.NewEncoder(iw)
				err = e.Encode(info)
				if err != nil {
					log.Fatal(err)
				}
			}()
		}

	},
}

func init() {
	packCmd.AddCommand(packOldManga)
	rootCmd.AddCommand(packCmd)
}
