package category

type Definition struct {
	ID   int64
	Slug string
}

const (
	NovelID    int64 = 1
	GuideID    int64 = 2
	FeedbackID int64 = 3
)

var definitions = [...]Definition{
	{ID: NovelID, Slug: "novel"},
	{ID: GuideID, Slug: "guide"},
	{ID: FeedbackID, Slug: "feedback"},
}

func List() []Definition {
	return append([]Definition(nil), definitions[:]...)
}

func FindByID(id int64) (Definition, bool) {
	for _, item := range definitions {
		if item.ID == id {
			return item, true
		}
	}
	return Definition{}, false
}

func FindBySlug(slug string) (Definition, bool) {
	for _, item := range definitions {
		if item.Slug == slug {
			return item, true
		}
	}
	return Definition{}, false
}
