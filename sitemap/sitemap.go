package sitemap

type Sitemap struct {
	XMLName string `xml:"http://www.sitemaps.org/schemas/sitemap/0.9 urlset"`
	URLs    []URL
}

type URL struct {
	XMLName string `xml:"url"`
	Loc     string `xml:"loc"`
}

// <?xml version="1.0" encoding="UTF-8"?>
// <urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
//    <url>
//       <loc>http://www.example.com/</loc>
//       <lastmod>2005-01-01</lastmod>
//       <changefreq>monthly</changefreq>
//       <priority>0.8</priority>
//    </url>
// </urlset>
