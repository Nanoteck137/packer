package cli

import (
	"archive/zip"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/gosimple/slug"
	"github.com/kr/pretty"
	"github.com/nanoteck137/packer/metadata"
	"github.com/nanoteck137/packer/utils"
	"github.com/spf13/cobra"
)

var packCmd = &cobra.Command{
	Use: "pack",
}

type MangaInfoChapter struct {
	Index int      `json:"index"`
	Name  string   `json:"name"`
	Pages []string `json:"pages"`
}

type MangaInfo struct {
	Title    string             `json:"title"`
	Cover    string             `json:"cover"`
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

type ExtraInfo struct {
	MalId     string
	AnilistId string
}

func createSeries(info MangaInfo, extra ExtraInfo, cover string, out string) error {
	dname, err := os.MkdirTemp("", "packer-series")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(dname)

	large := path.Join(dname, "cover-large.png")
	err = utils.CreateResizedImage(cover, large, 360, 480)
	if err != nil {
		return err
	}

	medium := path.Join(dname, "cover-medium.png")
	err = utils.CreateResizedImage(cover, medium, 270, 360)
	if err != nil {
		return err
	}

	small := path.Join(dname, "cover-small.png")
	err = utils.CreateResizedImage(cover, small, 180, 240)
	if err != nil {
		return err
	}

	seriesInfo := metadata.SeriesInfo{
		Name:      info.Title,
		Type:      metadata.SeriesTypeManga,
		MalId:     extra.MalId,
		AnilistId: extra.AnilistId,
		Cover: metadata.SeriesInfoCover{
			Original: path.Base(cover),
			Small:    path.Base(small),
			Medium:   path.Base(medium),
			Large:    path.Base(large),
		},
	}

	name := slug.Make(info.Title)

	f, err := os.OpenFile(path.Join(out, name+".sws"), os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	w := zip.NewWriter(f)
	defer w.Close()

	copyFileToZip := func(file string) error {
		h := &zip.FileHeader{
			Name: path.Base(file),
		}

		w, err := w.CreateHeader(h)
		if err != nil {
			return err
		}

		src, err := os.Open(file)
		if err != nil {
			return err
		}
		defer src.Close()

		_, err = io.Copy(w, src)
		if err != nil {
			return err
		}

		return nil
	}

	err = copyFileToZip(large)
	if err != nil {
		return err
	}

	err = copyFileToZip(medium)
	if err != nil {
		return err
	}

	err = copyFileToZip(small)
	if err != nil {
		return err
	}

	h := &zip.FileHeader{
		Name: "info.json",
	}

	iw, err := w.CreateHeader(h)
	if err != nil {
		return err
	}

	e := json.NewEncoder(iw)
	e.SetIndent("", "  ")
	err = e.Encode(seriesInfo)
	if err != nil {
		return err
	}

	return nil
}

var packCbz = &cobra.Command{
	Use:  "cbz <INPUT> <OUTPUT>",
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		input := args[0]
		out := args[1]

		fmt.Printf("input: %v\n", input)
		fmt.Printf("out: %v\n", out)

		err := os.MkdirAll(out, 0755)
		if err != nil {
			log.Fatal("Failed to create out dir", err)
		}

		r, err := zip.OpenReader(input)
		if err != nil {
			log.Fatal("Failed to open input", err)
		}

		var comicInfoFile *zip.File
		var pagesFound []*zip.File

		for _, f := range r.File {
			fmt.Printf("f.Name: %v\n", f.Name)

			if f.Name == "ComicInfo.xml" {
				comicInfoFile = f
				continue
			}

			ext := path.Ext(f.Name)
			// TODO(patrik): Support gif and other image formats by
			// converting them to pngs
			switch ext {
			case ".jpeg", ".jpg", ".png":
				pagesFound = append(pagesFound, f)
			default:
				log.Println("Warning: Unsupported extention found:", ext)
			}
		}

		_ = comicInfoFile

		type ComicInfo struct {
			Title  string
			Series string
			Manga  string
		}

		readComicInfo := func() (ComicInfo, error) {
			if comicInfoFile == nil {
				return ComicInfo{}, errors.New("No ComicInfo.xml found")
			}

			r, err := comicInfoFile.Open()
			if err != nil {
				return ComicInfo{}, nil
			}
			defer r.Close()

			var info ComicInfo
			d := xml.NewDecoder(r)
			err = d.Decode(&info)
			if err != nil {
				return ComicInfo{}, nil
			}

			return info, nil
		}

		comicInfo, err := readComicInfo()
		if err != nil {
			log.Fatal(err)
		}

		name := strings.TrimSpace(comicInfo.Title)

		pretty.Println(comicInfo)

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
		for i, page := range pagesFound {
			// TODO(patrik): Check ext for jpeg png
			ext := path.Ext(page.Name)
			n := strconv.Itoa(i) + ext
			pages = append(pages, n)

			h := &zip.FileHeader{
				Name: n,
			}

			w, err := w.CreateHeader(h)
			if err != nil {
				log.Fatal(err)
			}

			src, err := page.Open()
			if err != nil {
				log.Fatal(err)
			}
			defer src.Close()

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

		cf := pagesFound[0]
		dst := path.Join(dname, "cover-tmp"+path.Ext(cf.Name))
		{
			r, err := cf.Open()
			if err != nil {
				log.Fatal(err)
			}
			defer r.Close()

			f, err := os.Create(dst)
			if err != nil {
				log.Fatal(err)
			}
			defer f.Close()

			_, err = io.Copy(f, r)
			if err != nil {
				log.Fatal(err)
			}
		}

		coverDst := path.Join(dname, "cover.png")
		err = utils.CreateResizedImage(dst, coverDst, 80, 112)
		if err != nil {
			log.Fatal(err)
		}

		h := &zip.FileHeader{
			Name: "cover.png",
		}

		cw, err := w.CreateHeader(h)
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

		info := metadata.EntryInfo{
			Name:           name,
			Series:         strings.TrimSpace(comicInfo.Series),
			IsManga:        comicInfo.Manga != "",
			PreferVertical: false,
			Cover:          "cover.png",
			Pages:          pages,
		}

		h = &zip.FileHeader{
			Name: "info.json",
		}

		iw, err := w.CreateHeader(h)
		if err != nil {
			log.Fatal(err)
		}

		e := json.NewEncoder(iw)
		e.SetIndent("", "  ")
		err = e.Encode(info)
		if err != nil {
			log.Fatal(err)
		}

	},
}

var packOldManga = &cobra.Command{
	Use:  "old-manga <BASE> <OUTPUT>",
	Args: cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		base := args[0]
		out := args[1]

		malId, _ := cmd.Flags().GetString("mal")
		anilistId, _ := cmd.Flags().GetString("anilist")

		err := os.MkdirAll(out, 0755)
		if err != nil {
			log.Fatal("Failed to create out dir", err)
		}

		// NOTE(patrik):
		//   <NAME>.sw - Sewaddle Entry (chapters)
		//     info.json
		//     *PAGES*.jpg|png
		//     cover.png 80x112

		mangaInfo, err := ReadMangaInfo(path.Join(base, "manga.json"))
		if err != nil {
			log.Fatal(err)
		}

		extra := ExtraInfo{
			MalId:     malId,
			AnilistId: anilistId,
		}

		err = createSeries(mangaInfo, extra, path.Join(base, "images", mangaInfo.Cover), out)
		if err != nil {
			log.Fatal(err)
		}

		for _, c := range mangaInfo.Chapters {
			func() {
				p := path.Join(base, "chapters", strconv.Itoa(c.Index))

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

				info := metadata.EntryInfo{
					Name:           c.Name,
					Series:         mangaInfo.Title,
					IsManga:        true,
					PreferVertical: false,
					Cover:          "cover.png",
					Pages:          pages,
				}

				h := &zip.FileHeader{
					Name: "info.json",
				}

				iw, err := w.CreateHeader(h)
				if err != nil {
					log.Fatal(err)
				}

				e := json.NewEncoder(iw)
				e.SetIndent("", "  ")
				err = e.Encode(info)
				if err != nil {
					log.Fatal(err)
				}
			}()
		}

	},
}

func init() {
	packOldManga.Flags().String("mal", "", "Set MyAnimeList ID")
	packOldManga.Flags().String("anilist", "", "Set AniList ID")

	packCmd.AddCommand(packCbz)
	packCmd.AddCommand(packOldManga)

	rootCmd.AddCommand(packCmd)
}
