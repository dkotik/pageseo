package sitemap

import (
	"os"
	"testing"

	"github.com/dkotik/pageseo"
)

func TestSiteMapValidation(t *testing.T) {
	loader := pageseo.NewFS(os.DirFS("testdata"))
	Test(loader, "single.xml")(t)
}
