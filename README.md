# Technet Blog Feed Finder

A couple of years back, a post landed on [IT Pro Today](https://www.itprotoday.com/windows-10/resource-rss-feed-lists-microsoft-tech-community-sites) that linked to an OPML file for all of Microsoft's Technical Community blogs. The problem became that the link eventually died, and the resource was never updated.

There are over [110 different communities](https://techcommunity.microsoft.com/t5/communities/ct-p/communities), and something like [175 blogs](https://techcommunity.microsoft.com/t5/custom/page/page-id/Blogs) that MS uses to communicate what's going on with their products and platforms.

Keeping up with the additions (and removals?) of the blogs gets to get a chore.  That's what this project does.

## How it works

This is me learning Go, so it's a bit rough around the edges, but it generally works by:

- Reading the Tech Community's Blog page and finding the list of blogs (using [GoQuery](https://github.com/PuerkitoBio/goquery))
- Checking each blog's page for the RSS feed and the category (which I think maps to the Community)
- Generate an OPML RSS feed.

## Observations

I noticed that the RSS feed that Microsoft would give you is "pretty" - something like `https://techcommunity.microsoft.com/plugins/custom/microsoft/o365/custom-blog-rss?board=AccessBlog&size=25`, but the feed provided in the metadata of the page is `https://techcommunity.microsoft.com/gxcuf89792/rss/board?board.id=AccessBlog`. I've chosen to collect the feed URL from the page metadata, instead of the constructed URL. Hopefully MS doesn't run off and change it.
