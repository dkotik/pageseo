/*
Package slug provides functions for generating optimized URL slugs from strings.

Reference:

- https://neilpatel.com/blog/seo-urls/
- https://cseo.com/blog/seo-stop-words/
*/
package slug

import "errors"

func New(text string) (string, error) {
	return "", errors.New("not implemented")
}

/*
// Page titles should be short and concise because search engines read no more than 55 to 60 characters, including spaces.

// The bottom line is that you want to condense the essence of your content into roughly three to five words and try to use a max of 60 characters.
// Use hyphens
// Use lowercase letters
// Use a max of two path folds per URL
// First of all, you should definitely still include keywords in your URL. According to Brian Dean of Backlinko and John Lincoln, CEO of Ignite Visibility, you should aim for one or two keywords per URL.
// If you feel like you need to include a stop word for your URL to make sense and more readable then go ahead and include it. If it makes it easier for people to read, then that’s usually your best option.

export const validator = path => {
  const pathWords = new Set();
  // Never repeat your keywords (or any words for that matter) in a whole URL, not just the slug. So if URL starts with /learning-center/ never use words "learning" or "center" in the rest of it.
  for (const fold of path.split("/")) {
    for (const word of path.split("-")) {
      pathWords.add(word);
    }
  }

  return slug => {
    try {
      if (typeof slug !== "string")
        throw new Error("provided slug is not a string");
      if (typeof path !== "string")
        throw new Error("provided path is not a string");
      if (!slug)
        throw new Error(`"slug" field is required but missing; it should be injected using a Markdown plugin or provided from MDX.file field matched using matchSlug()`);

      // if (!path.match(reMatchSlug))
      if (slug.length + path.length > 60)
        throw new Error("slug must not exceed 60 characters in length");

      const words = slug.split("-");
      if (words.length < 3)
        throw new Error("slug must contain at least 3 words");

      const duplicates = new Set();
      for (const word of words) {
        if (word.legth < 2) {
          throw new Error(
            `word "${word}" rejected: slug words must contain at least 2 letters`
          );
        }

        if (pathWords.has(word)) {
          throw new Error(`word "${word}" rejected: the same word is in path`);
        }

        if (duplicates.has(word)) {
          throw new Error(`word "${word}" rejected: it occurs more than once`);
        } else {
          duplicates.add(word);
        }
        if (word in stopWords)
          throw new Error(`word "${word}" is rejected: it is a SEO stop word`);
        if (!reValidWord.test(word))
          throw new Error(
            `word "${word}" is rejected: it must contain only lowercase letters or numbers and must begin with a letter, or be a year in 20XX format`
          );
      }
    } catch (e) {
      throw new Error(
        `could not validate provided URL slug "${slug}" behind path "${path}": ${e}"`
      );
    }
  };
};
*/
