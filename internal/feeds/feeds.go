package feeds

import "strings"

// Feed represents a curated RSS feed from the awesome-rss-feeds database.
type Feed struct {
	Name        string
	URL         string
	Description string
}

// Category groups feeds by topic area.
type Category struct {
	Name  string
	Feeds []Feed
}

// Categories contains all curated RSS feeds organized by category.
var Categories = []Category{
	{
		Name: "Android Development",
		Feeds: []Feed{
			{Name: "Android - Buffer Resources", URL: "https://buffer.com/resources/android/rss/", Description: "In-depth ideas and guides to social media & online marketing strategy, publis..."},
			{Name: "Android Developers", URL: "https://www.youtube.com/feeds/videos.xml?user=androiddevelopers", Description: "Android Developers"},
			{Name: "Android Developers - Medium", URL: "https://medium.com/feed/androiddevelopers", Description: "The official Android Developers publication on Medium - Medium"},
			{Name: "Android Developers Backstage", URL: "http://feeds.feedburner.com/blogspot/androiddevelopersbackstage", Description: "Android Backstage, a podcast by and for Android developers. Hosted by develop..."},
			{Name: "Android Developers Blog", URL: "http://feeds.feedburner.com/blogspot/hsDu", Description: "An Open Handset Alliance Project."},
			{Name: "Android Weekly Archive Feed", URL: "https://us2.campaign-archive.com/feed?u=887caf4f48db76fd91e20a06d&id=4eb677ad19", Description: "Android Weekly Archive Feed"},
			{Name: "Android in Instagram Engineering on Medium", URL: "https://instagram-engineering.com/feed/tagged/android", Description: "Latest stories tagged with Android in Instagram Engineering on Medium"},
			{Name: "Android in MindOrks on Medium", URL: "https://medium.com/feed/mindorks/tagged/android", Description: "Latest stories tagged with Android in MindOrks on Medium"},
			{Name: "Android in The Airbnb Tech Blog on Medium", URL: "https://medium.com/feed/airbnb-engineering/tagged/android", Description: "Latest stories tagged with Android in The Airbnb Tech Blog on Medium"},
			{Name: "Dan Lew Codes", URL: "https://blog.danlew.net/rss/", Description: "Thoughts on life, the universe and the mystery of it all; but actually mostly..."},
			{Name: "Developing Android Apps", URL: "https://reddit.com/r/androiddev.rss", Description: "News for Android developers with the who, what, where, when and how of the An..."},
			{Name: "Fragmented - The Software Podcast", URL: "https://feeds.simplecast.com/LpAGSLnY", Description: "The Fragmented Podcast is a podcast for Software Developers hosted by Donn Fe..."},
			{Name: "Handstand Sam", URL: "https://handstandsam.com/feed/", Description: "Sam Edwards - Handstands, Travel, Android & Web"},
			{Name: "Jake Wharton", URL: "https://jakewharton.com/atom.xml", Description: "Jake Wharton"},
			{Name: "JetBrains News | JetBrains Blog", URL: "https://blog.jetbrains.com/blog/feed", Description: "Developer Tools for Professionals and Teams"},
			{Name: "Joe Birch", URL: "https://joebirch.co/feed", Description: "Speaker, Educator and GDE for Android, Google Assistant & Flutter"},
			{Name: "Kotlin", URL: "https://www.youtube.com/feeds/videos.xml?playlist_id=PLQ176FUIyIUa6SChjajjVc-LMzxWiz6dy", Description: "Kotlin"},
			{Name: "Kt. Academy - Medium", URL: "https://blog.kotlin-academy.com/feed", Description: "Blog with mission to simplify Kotlin learning - Medium"},
			{Name: "OkKotlin", URL: "https://okkotlin.com/rss.xml", Description: "A premier blog on Kotlin."},
			{Name: "ProAndroidDev - Medium", URL: "https://proandroiddev.com/feed", Description: "The latest posts from Android Professionals and Google Developer Experts. - M..."},
			{Name: "Public Object", URL: "https://publicobject.com/rss/", Description: "Jesse Wilson on programming."},
			{Name: "Saket Narayan", URL: "https://saket.me/feed/", Description: "A wild developer appears. He works at <a href="},
			{Name: "Styling Android", URL: "http://feeds.feedburner.com/StylingAndroid", Description: "A technical guide to improving the UI and UX of Android apps"},
			{Name: "Talking Kotlin", URL: "https://feeds.soundcloud.com/users/soundcloud:users:280353173/sounds.rss", Description: "A bimonthly podcast that covers the Kotlin programming language by JetBrains,..."},
			{Name: "The Android Arsenal", URL: "https://feeds.feedburner.com/Android_Arsenal", Description: "A categorized directory of libraries and tools for Android"},
			{Name: "Zac Sweers", URL: "https://www.zacsweers.dev/rss/", Description: "Life, travel, code, and the whitespace in between."},
			{Name: "Zarah Dominguez", URL: "https://zarah.dev/feed.xml", Description: "An Android Love Affair"},
			{Name: "chRyNaN Codes", URL: "https://chrynan.codes/rss/", Description: "Software Development Blog"},
			{Name: "droidcon NYC", URL: "https://www.youtube.com/feeds/videos.xml?channel_id=UCSLXy31j2Z0sdDeeAX5JpPw", Description: "droidcon NYC"},
			{Name: "droidcon SF", URL: "https://www.youtube.com/feeds/videos.xml?channel_id=UCKubKoe1CBw_-n_GXetEQbg", Description: "droidcon SF"},
			{Name: "goobar", URL: "https://goobar.io/feed", Description: "dream / learn / create"},
			{Name: "zsmb.co", URL: "https://zsmb.co/index.xml", Description: "Recent content on zsmb.co"},
		},
	},
	{
		Name: "Android",
		Feeds: []Feed{
			{Name: "All About Android (Audio)", URL: "https://feeds.twit.tv/aaa.xml", Description: "All About Android delivers everything you want to know about Android each wee..."},
			{Name: "Android", URL: "https://blog.google/products/android/rss", Description: "Android"},
			{Name: "Android", URL: "https://www.reddit.com/r/android/.rss", Description: "Android news, reviews, tips, and discussions about rooting, tutorials, and ap..."},
			{Name: "Android Authority", URL: "https://www.androidauthority.com/feed", Description: "Android News, Reviews, How To"},
			{Name: "Android Authority", URL: "https://www.youtube.com/feeds/videos.xml?user=AndroidAuthority", Description: "Android Authority"},
			{Name: "Android Central - Android Forums, News, Reviews, Help and Android Wallpapers", URL: "http://feeds.androidcentral.com/androidcentral", Description: "Android Central - Android Forums, News, Reviews, Help and Android Wallpapers"},
			{Name: "Android Central Podcast", URL: "http://feeds.feedburner.com/AndroidCentralPodcast", Description: "The Android and Google Podcast for Everyone. Hosted by Daniel Bader."},
			{Name: "Android Community", URL: "https://androidcommunity.com/feed/", Description: "Android News, Reviews, and Apps"},
			{Name: "Android Police – Android news, reviews, apps, games, phones, tablets", URL: "http://feeds.feedburner.com/AndroidPolice", Description: "Looking after everything Android"},
			{Name: "AndroidGuys", URL: "https://www.androidguys.com/feed", Description: "Android news and opinion"},
			{Name: "Cult of Android", URL: "https://www.cultofandroid.com/feed", Description: "Breaking news for Android fans"},
			{Name: "Cyanogen Mods", URL: "https://www.cyanogenmods.org/feed", Description: "Mod APK | GCam | LineageOS"},
			{Name: "Droid Life", URL: "https://www.droid-life.com/feed", Description: "Opinionated Android news."},
			{Name: "GSMArena.com - Latest articles", URL: "https://www.gsmarena.com/rss-news-reviews.php3", Description: "GSMArena.com is the ultimate resource for GSM handset information. This feed ..."},
			{Name: "Phandroid", URL: "http://feeds2.feedburner.com/AndroidPhoneFans", Description: "Android Phone News, Rumors, Reviews, Apps, Forums & More!"},
			{Name: "TalkAndroid", URL: "http://feeds.feedburner.com/AndroidNewsGoogleAndroidForums", Description: "The latest in android news, rumours, and updates, including device news and a..."},
			{Name: "xda-developers", URL: "https://data.xda-developers.com/portal-feed", Description: "Android and Windows Phone Development Community"},
		},
	},
	{
		Name: "Apple",
		Feeds: []Feed{
			{Name: "9to5Mac", URL: "https://9to5mac.com/feed", Description: "Apple News & Mac Rumors Breaking All Day"},
			{Name: "Apple", URL: "https://www.youtube.com/feeds/videos.xml?user=Apple", Description: "Apple"},
			{Name: "Apple Newsroom", URL: "https://www.apple.com/newsroom/rss-feed.rss", Description: "Apple Newsroom"},
			{Name: "AppleInsider News", URL: "https://appleinsider.com/rss/news/", Description: "AppleInsider News Feed"},
			{Name: "Cult of Mac", URL: "https://www.cultofmac.com/feed", Description: "Tech and culture through an Apple lens"},
			{Name: "Daring Fireball", URL: "https://daringfireball.net/feeds/main", Description: "By John Gruber"},
			{Name: "MacRumors", URL: "https://www.youtube.com/feeds/videos.xml?user=macrumors", Description: "MacRumors"},
			{Name: "MacRumors: Mac News and Rumors - Mac Blog", URL: "http://feeds.macrumors.com/MacRumors-Mac", Description: "Apple, iPhone, iPad, Mac News and Rumors"},
			{Name: "MacStories", URL: "https://www.macstories.net/feed", Description: "Apple news, app reviews, and stories by Federico Viticci and friends."},
			{Name: "Macworld.com", URL: "https://www.macworld.com/index.rss", Description: "Macworld is your best source for all things Apple. We give you the scoop on w..."},
			{Name: "Marco.org", URL: "https://marco.org/rss", Description: "I’m Marco Arment, creator of Overcast, technology podcaster and writer, and c..."},
			{Name: "OS X Daily", URL: "http://feeds.feedburner.com/osxdaily", Description: "News, tips, software, reviews, and more for Mac OS X, iPhone, iPad"},
			{Name: "The Loop", URL: "https://www.loopinsight.com/feed", Description: "Making Sense of Technology"},
			{Name: "The unofficial Apple community", URL: "https://www.reddit.com/r/apple/.rss", Description: "An unofficial community to discuss Apple devices and software, including news..."},
			{Name: "iMore - The #1 iPhone, iPad, and iPod touch blog", URL: "http://feeds.feedburner.com/TheiPhoneBlog", Description: "More news and rumors, more help and how-tos, more app and accessory reviews, ..."},
			{Name: "r/iPhone", URL: "https://www.reddit.com/r/iphone/.rss", Description: "Reddit’s corner for iPhone lovers (or those who mildly enjoy it...)"},
		},
	},
	{
		Name: "Architecture",
		Feeds: []Feed{
			{Name: "A Daily Dose of Architecture Books", URL: "http://feeds.feedburner.com/archidose", Description: "(Almost) daily reviews of architecture books"},
			{Name: "ArchDaily", URL: "http://feeds.feedburner.com/Archdaily", Description: "ArchDaily | Broadcasting Architecture Worldwide"},
			{Name: "Archinect - News", URL: "https://archinect.com/feed/1/news", Description: "Archinect - News"},
			{Name: "Architectural Digest", URL: "https://www.architecturaldigest.com/feed/rss", Description: "The latest from www.architecturaldigest.com"},
			{Name: "Architectural Digest", URL: "https://www.youtube.com/feeds/videos.xml?user=ArchitecturalDigest", Description: "Architectural Digest"},
			{Name: "Architecture", URL: "https://www.reddit.com/r/architecture/.rss", Description: "A community for students, professionals, and lovers of architecture."},
			{Name: "Architecture – Dezeen", URL: "https://www.dezeen.com/architecture/feed/", Description: "architecture and design magazine"},
			{Name: "CONTEMPORIST", URL: "https://www.contemporist.com/feed/", Description: "Contemporist features great ideas from the world of design, architecture, int..."},
			{Name: "Comments on:", URL: "https://inhabitat.com/architecture/feed/", Description: "Green design & innovation for a better world"},
			{Name: "Design MilkArchitecture – Design Milk", URL: "https://design-milk.com/category/architecture/feed/", Description: "Dedicated to modern design"},
			{Name: "Journal", URL: "https://architizer.wpengine.com/feed/", Description: "Inspiration and Tools for Architects"},
			{Name: "Living Big In A Tiny House", URL: "https://www.youtube.com/feeds/videos.xml?user=livingbigtinyhouse", Description: "Living Big In A Tiny House"},
			{Name: "The Architect’s Newspaper", URL: "https://archpaper.com/feed", Description: "The most authoritative voice on architecture and design in the United States."},
			{Name: "architecture – designboom | architecture & design magazine", URL: "https://www.designboom.com/architecture/feed/", Description: "designboom magazine | your first source for architecture, design & art news"},
		},
	},
	{
		Name: "Beauty",
		Feeds: []Feed{
			{Name: "Beauty - ELLE", URL: "https://www.elle.com/rss/beauty.xml/", Description: "Beauty - ELLE"},
			{Name: "Beauty - Fashionista", URL: "https://fashionista.com/.rss/excerpt/beauty", Description: "Beauty - Fashionista"},
			{Name: "Beauty – Indian Fashion Blog", URL: "https://www.fashionlady.in/category/beauty-tips/feed", Description: "Beauty – Indian Fashion Blog"},
			{Name: "Blog – The Beauty Brains", URL: "https://thebeautybrains.com/blog/feed/", Description: "Real scientists answer your beauty questions"},
			{Name: "DORÉ", URL: "https://www.wearedore.com/feed", Description: "DORÉ"},
			{Name: "From Head To Toe", URL: "http://feeds.feedburner.com/frmheadtotoe", Description: "A makeup and beauty blog by Jen"},
			{Name: "Into The Gloss - Beauty Tips, Trends, And Product Reviews", URL: "https://feeds.feedburner.com/intothegloss/oqoU", Description: "The best in beauty tips, makeup tutorials, product reviews, and techniques fr..."},
			{Name: "POPSUGAR Beauty", URL: "https://www.popsugar.com/beauty/feed", Description: "Top Beauty Stories"},
			{Name: "Refinery29", URL: "https://www.refinery29.com/beauty/rss.xml", Description: "Refinery29"},
			{Name: "THE YESSTYLIST – Asian Fashion Blog – brought to you by YesStyle.com", URL: "https://www.yesstyle.com/blog/category/the-beauty-blog/feed/", Description: "THE YESSTYLIST – Asian Fashion Blog – brought to you by YesStyle.com"},
			{Name: "The Beauty Look Book", URL: "https://thebeautylookbook.com/feed", Description: "Beauty Blog, Reviews + Makeup Looks"},
		},
	},
	{
		Name: "Books",
		Feeds: []Feed{
			{Name: "A year of reading the world", URL: "https://ayearofreadingtheworld.com/feed/", Description: "196 countries, countless stories..."},
			{Name: "Aestas Book Blog", URL: "https://aestasbookblog.com/feed/", Description: "Romance book reviews. Reviews of books that make my heart race, have a beauti..."},
			{Name: "BOOK RIOT", URL: "https://bookriot.com/feed/", Description: "Book Recommendations and Reviews"},
			{Name: "Kirkus Reviews", URL: "https://www.kirkusreviews.com/feeds/rss/", Description: "Latest book reviews from Kirkus Reviews."},
			{Name: "Page Array – NewInBooks", URL: "https://www.newinbooks.com/feed/", Description: "Answering the Age old question - What are you reading?"},
			{Name: "So many books, so little time", URL: "https://reddit.com/r/books/.rss", Description: "This is a moderated subreddit. It is our intent and purpose to foster and enc..."},
			{Name: "Wokeread", URL: "https://wokeread.home.blog/feed/", Description: "Books any time any day."},
		},
	},
	{
		Name: "Business & Economy",
		Feeds: []Feed{
			{Name: "All News", URL: "https://www.investing.com/rss/news.rss", Description: "All News"},
			{Name: "Bloomberg Quicktake", URL: "https://www.youtube.com/feeds/videos.xml?user=Bloomberg", Description: "Bloomberg Quicktake"},
			{Name: "Breaking News on Seeking Alpha", URL: "https://seekingalpha.com/market_currents.xml", Description: "© seekingalpha.com. Use of this feed is limited to personal, non-commercial u..."},
			{Name: "Business Insider", URL: "https://www.youtube.com/feeds/videos.xml?user=businessinsider", Description: "Business Insider"},
			{Name: "Duct Tape Marketing", URL: "https://ducttape.libsyn.com/rss", Description: "Interviews with authors, experts and thought leaders sharing business marketi..."},
			{Name: "Economic Times", URL: "https://economictimes.indiatimes.com/rssfeedsdefault.cms", Description: "The Economic Times: Latest business, finance, markets, stocks, company news f..."},
			{Name: "Forbes - Business", URL: "https://www.forbes.com/business/feed/", Description: "Forbes - Business"},
			{Name: "Fortune", URL: "https://fortune.com/feed", Description: "Fortune 500 Daily & Breaking Business News"},
			{Name: "HBR IdeaCast", URL: "http://feeds.harvardbusiness.org/harvardbusiness/ideacast", Description: "A weekly podcast featuring the leading thinkers in business and management."},
			{Name: "Home Page", URL: "https://www.business-standard.com/rss/home_page_top_stories.rss", Description: "Business News: Get latest stock share market news, financial news, economy ne..."},
			{Name: "How I Built This with Guy Raz", URL: "https://feeds.npr.org/510313/podcast.xml", Description: "Guy Raz dives into the stories behind some of the world's best known companie..."},
			{Name: "Startup Stories - Mixergy", URL: "https://feeds.feedburner.com/Mixergy-main-podcast", Description: "Business tips for startups by proven entrepreneurs"},
			{Name: "The Blog of Author Tim Ferriss", URL: "https://tim.blog/feed/", Description: "Tim Ferriss's 4-Hour Workweek and Lifestyle Design Blog. Tim is an author of ..."},
			{Name: "The Growth Show", URL: "http://thegrowthshow.hubspot.libsynpro.com/", Description: "It’s never been easier to start a business, but it’s even harder to succeed. ..."},
			{Name: "US Top News and Analysis", URL: "https://www.cnbc.com/id/100003114/device/rss/rss.html", Description: "CNBC is the world leader in business news and real-time financial market cove..."},
			{Name: "Yahoo Finance", URL: "https://finance.yahoo.com/news/rssindex", Description: "At Yahoo Finance, you get free stock quotes, up-to-date news, portfolio manag..."},
		},
	},
	{
		Name: "Cars",
		Feeds: []Feed{
			{Name: "Autoblog", URL: "https://www.autoblog.com/rss.xml", Description: "Autoblog"},
			{Name: "Autocar India - All Bike Reviews", URL: "https://www.autocarindia.com/RSS/rss.ashx?type=all_bikes", Description: "Car and Bike news, reviews and videos."},
			{Name: "Autocar India - All Car Reviews", URL: "https://www.autocarindia.com/RSS/rss.ashx?type=all_cars", Description: "Car and Bike news, reviews and videos."},
			{Name: "Autocar India - News", URL: "https://www.autocarindia.com/RSS/rss.ashx?type=News", Description: "Car and Bike news, reviews and videos."},
			{Name: "Autocar RSS Feed", URL: "https://www.autocar.co.uk/rss", Description: "Welcome to nirvana for car enthusiasts. You have just entered the online home..."},
			{Name: "BMW BLOG", URL: "https://feeds.feedburner.com/BmwBlog", Description: "BMW News, Reviews, Test Drives, Photos And Videos"},
			{Name: "Bike EXIF", URL: "https://www.bikeexif.com/feed", Description: "The world's most exciting custom motorcycles, from cafe racers to bobbers to ..."},
			{Name: "Car Body Design", URL: "https://www.carbodydesign.com/feed/", Description: "Car Design Resources, News and Tutorials"},
			{Name: "Carscoops", URL: "https://www.carscoops.com/feed/", Description: "Breaking Car News, Scoops & Reviews"},
			{Name: "Formula 1", URL: "https://www.reddit.com/r/formula1/.rss", Description: "The best independent Formula 1 community anywhere. News, stories and discussi..."},
			{Name: "Jalopnik", URL: "https://jalopnik.com/rss", Description: "Kinja RSS"},
			{Name: "Latest Content - Car and Driver", URL: "https://www.caranddriver.com/rss/all.xml/", Description: "Latest Content - Car and Driver"},
			{Name: "Petrolicious", URL: "https://petrolicious.com/feed", Description: "Petrolicious is a leading automotive lifestyle brand providing world class sh..."},
			{Name: "Section Page News - Automotive News", URL: "http://feeds.feedburner.com/autonews/AutomakerNews", Description: "Section Page news and features from Automotive News"},
			{Name: "Section Page News - Automotive News", URL: "http://feeds.feedburner.com/autonews/EditorsPicks", Description: "Section Page news and features from Automotive News"},
			{Name: "Speedhunters", URL: "http://feeds.feedburner.com/speedhunters", Description: "Global Car Culture Since 2008"},
			{Name: "The Truth About Cars", URL: "https://www.thetruthaboutcars.com/feed/", Description: "The Truth About Cars is dedicated to providing candid, unbiased automobile re..."},
			{Name: "The best vintage and classic cars for sale online | Bring a Trailer", URL: "https://bringatrailer.com/feed/", Description: "Barn finds, rally cars, and needles in the haystack."},
		},
	},
	{
		Name: "Cricket",
		Feeds: []Feed{
			{Name: "BBC Sport - Cricket", URL: "http://feeds.bbci.co.uk/sport/cricket/rss.xml", Description: "BBC Sport - Cricket"},
			{Name: "Can't Bowl Can't Throw Cricket Show", URL: "http://feeds.feedburner.com/cantbowlcantthrow", Description: "The second season of the Can't Bowl Can't Throw Cricket Show, in which we tal..."},
			{Name: "Cricbuzz", URL: "https://www.youtube.com/feeds/videos.xml?channel_id=UCSRQXk5yErn4e14vN76upOw", Description: "Cricbuzz"},
			{Name: "Cricket", URL: "https://www.reddit.com/r/Cricket/.rss", Description: "News, banter and occasional serious discussion on the great game."},
			{Name: "Cricket Unfiltered", URL: "https://rss.acast.com/cricket-unfiltered", Description: "<p>If you love cricket then this podcast is for you. Get closer to the game t..."},
			{Name: "Cricket news from ESPN Cricinfo.com", URL: "http://www.espncricinfo.com/rss/content/story/feeds/0.xml", Description: "Visit Cricinfo.com for up-to-the-minute cricket news, breaking cricket news, ..."},
			{Name: "Cricket | The Guardian", URL: "https://www.theguardian.com/sport/cricket/rss", Description: "Latest news and features from theguardian.com, the world's leading liberal voice"},
			{Name: "Cricket – The Roar", URL: "https://www.theroar.com.au/cricket/feed/", Description: "Australia's Biggest Sporting Debate"},
			{Name: "England & Wales Cricket Board", URL: "https://www.youtube.com/feeds/videos.xml?user=ecbcricket", Description: "England & Wales Cricket Board"},
			{Name: "NDTV Sports - Cricket", URL: "http://feeds.feedburner.com/ndtvsports-cricket", Description: "NDTV.com provides the latest information from and in-depth coverage of India ..."},
			{Name: "Pakistan Cricket", URL: "https://www.youtube.com/feeds/videos.xml?channel_id=UCiWrjBhlICf_L_RK5y6Vrxw", Description: "Pakistan Cricket"},
			{Name: "Sky Sports Cricket Podcast", URL: "https://www.spreaker.com/show/3387348/episodes/feed", Description: "Sky's pundits analyse and debate the big stories emerging from international ..."},
			{Name: "Sri Lanka Cricket", URL: "https://www.youtube.com/feeds/videos.xml?user=TheOfficialSLC", Description: "Sri Lanka Cricket"},
			{Name: "Stumped", URL: "https://podcasts.files.bbci.co.uk/p02gsrmh.rss", Description: "The weekly cricket show from BBC Sport in association with ABC and All India ..."},
			{Name: "Switch Hit Podcast", URL: "https://feeds.megaphone.fm/ESP9247246951", Description: "Lively debate and discussion about the game on and off the field with a focus..."},
			{Name: "Tailenders", URL: "https://podcasts.files.bbci.co.uk/p02pcb4w.rss", Description: "Greg James, Jimmy Anderson and Felix White with an alternative (and sometimes..."},
			{Name: "Test Match Special", URL: "https://podcasts.files.bbci.co.uk/p02nrsl2.rss", Description: "Insight and analysis from the Test Match Special team - including interviews ..."},
			{Name: "The Analyst Inside Cricket", URL: "http://rss.acast.com/theanalystinsidecricket", Description: "<p>Sharp analysis and opinion from the cricket world with Simon Hughes, The A..."},
			{Name: "The Grade Cricketer", URL: "https://rss.whooshkaa.com/rss/podcast/id/1308", Description: "Cricket is great if you're into things like wasted youth, failed relationship..."},
			{Name: "Wisden", URL: "https://www.wisden.com/feed", Description: "The independent voice of cricket"},
			{Name: "Wisden Cricket Weekly", URL: "http://feeds.soundcloud.com/users/soundcloud:users:341034518/sounds.rss", Description: "Join Mark Butcher, Yas Rana and the rest of the Wisden Cricket Weekly Podcast..."},
			{Name: "cricket.com.au", URL: "https://www.youtube.com/feeds/videos.xml?user=cricketaustraliatv", Description: "cricket.com.au"},
		},
	},
	{
		Name: "DIY",
		Feeds: []Feed{
			{Name: "A Beautiful Mess", URL: "https://abeautifulmess.com/feed", Description: "Crafts, Home Décor, Recipes"},
			{Name: "Apartment Therapy| Saving the world, one room at a time", URL: "https://www.apartmenttherapy.com/projects.rss", Description: "Saving the world, one room at a time"},
			{Name: "Blog – Hackaday", URL: "https://hackaday.com/blog/feed/", Description: "Fresh hacks every day"},
			{Name: "Centsational Style", URL: "https://centsationalstyle.com/feed/", Description: "Design, Decorating & DIY"},
			{Name: "Doityourself.com", URL: "https://www.doityourself.com/feed", Description: "Do it yourself home improvement and diy repair at Doityourself.com. Includes ..."},
			{Name: "Etsy Journal", URL: "https://blog.etsy.com/en/feed/", Description: "Ideas and inspiration for creative living."},
			{Name: "How-To Geek", URL: "https://www.howtogeek.com/feed/", Description: "We Explain Technology"},
			{Name: "IKEA Hackers", URL: "https://www.ikeahackers.net/feed", Description: "Clever ideas and hacks for your IKEA"},
			{Name: "MUO - Feed", URL: "https://www.makeuseof.com/feed/", Description: "MUO is your guide in modern tech. Learn how to make use of tech and gadgets a..."},
			{Name: "Oh Happy Day!", URL: "http://ohhappyday.com/feed/", Description: "Oh Happy Day is a party and lifestyle blog based in San Francisco. We love to..."},
			{Name: "WonderHowTo", URL: "https://www.wonderhowto.com/rss.xml", Description: "Hot new how-to videos and articles on WonderHowTo."},
		},
	},
	{
		Name: "Fashion",
		Feeds: []Feed{
			{Name: "Fashion - ELLE", URL: "https://www.elle.com/rss/fashion.xml/", Description: "Fashion - ELLE"},
			{Name: "Fashion | The Guardian", URL: "https://www.theguardian.com/fashion/rss", Description: "The latest fashion news, advice and comment from the Guardian"},
			{Name: "Fashion – Indian Fashion Blog", URL: "https://www.fashionlady.in/category/fashion/feed", Description: "Fashion – Indian Fashion Blog"},
			{Name: "FashionBeans Men's Fashion and Style Feed", URL: "https://www.fashionbeans.com/rss-feed/?category=fashion", Description: "A comprehensive guide to men's fashion and style, from the latest trends and ..."},
			{Name: "Fashionista", URL: "https://fashionista.com/.rss/excerpt/", Description: "Fashionista"},
			{Name: "NYT > Style", URL: "https://rss.nytimes.com/services/xml/rss/nyt/FashionandStyle.xml", Description: "NYT > Style"},
			{Name: "POPSUGAR Fashion", URL: "https://www.popsugar.com/fashion/feed", Description: "Top Fashion Stories"},
			{Name: "Refinery29", URL: "https://www.refinery29.com/fashion/rss.xml", Description: "Refinery29"},
			{Name: "THE YESSTYLIST – Asian Fashion Blog – brought to you by YesStyle.com", URL: "https://www.yesstyle.com/blog/category/trend-and-style/feed/", Description: "THE YESSTYLIST – Asian Fashion Blog – brought to you by YesStyle.com"},
			{Name: "Who What Wear", URL: "https://www.whowhatwear.com/rss", Description: "Who What Wear"},
		},
	},
	{
		Name: "Food",
		Feeds: []Feed{
			{Name: "101 Cookbooks", URL: "https://www.101cookbooks.com/feed", Description: "When you own over 100 cookbooks, it is time to stop buying, and start cooking..."},
			{Name: "Babish Culinary Universe", URL: "https://www.youtube.com/feeds/videos.xml?user=bgfilms", Description: "Babish Culinary Universe"},
			{Name: "Bon Appétit", URL: "https://www.youtube.com/feeds/videos.xml?user=BonAppetitDotCom", Description: "Bon Appétit"},
			{Name: "Chocolate & Zucchini", URL: "https://cnz.to/feed/", Description: "Simple Recipes from my Paris Kitchen"},
			{Name: "David Lebovitz", URL: "https://www.davidlebovitz.com/feed/", Description: "Baking and cooking recipes for everyone"},
			{Name: "Food52", URL: "http://feeds.feedburner.com/food52-TheAandMBlog", Description: "Eat thoughtfully, live joyfully."},
			{Name: "Green Kitchen Stories", URL: "https://greenkitchenstories.com/feed/", Description: "Healthy Vegetarian Recipes."},
			{Name: "How Sweet Eats", URL: "https://www.howsweeteats.com/feed/", Description: "For people who, like, totally love food."},
			{Name: "Joy the Baker", URL: "http://joythebaker.com/feed/", Description: "Joy the Baker"},
			{Name: "Kitchn | Inspiring cooks, nourishing homes", URL: "https://www.thekitchn.com/main.rss", Description: "Inspiring cooks, nourishing homes"},
			{Name: "Laura in the Kitchen", URL: "https://www.youtube.com/feeds/videos.xml?user=LauraVitalesKitchen", Description: "Laura in the Kitchen"},
			{Name: "Love and Olive Oil", URL: "https://www.loveandoliveoil.com/feed", Description: "Eat to Live. Cook to Love."},
			{Name: "NYT > Food", URL: "https://rss.nytimes.com/services/xml/rss/nyt/DiningandWine.xml", Description: "NYT > Food"},
			{Name: "Oh She Glows", URL: "https://ohsheglows.com/feed/", Description: "Vegan Recipes to Glow From The Inside Out"},
			{Name: "Serious Eats", URL: "https://www.youtube.com/feeds/videos.xml?user=SeriousEats", Description: "Serious Eats"},
			{Name: "Serious Eats: Recipes", URL: "http://feeds.feedburner.com/seriouseats/recipes", Description: "Our Favorite Recipes, Curated and Collected"},
			{Name: "Shutterbean", URL: "http://www.shutterbean.com/feed/", Description: "food, photography & inspiration"},
			{Name: "Skinnytaste", URL: "https://www.skinnytaste.com/feed/", Description: "Delicious Healthy Recipes Made with Real Food"},
			{Name: "Sprouted Kitchen", URL: "https://www.sproutedkitchen.com/home?format=rss", Description: "Sprouted Kitchen"},
			{Name: "Williams-Sonoma Taste", URL: "https://blog.williams-sonoma.com/feed/", Description: "Williams-Sonoma Taste"},
			{Name: "smitten kitchen", URL: "http://feeds.feedburner.com/smittenkitchen", Description: "Fearless cooking from a tiny NYC kitchen."},
		},
	},
	{
		Name: "Football",
		Feeds: []Feed{
			{Name: "EFL Championship", URL: "https://www.reddit.com/r/Championship/.rss?format=xml", Description: "Home of the EFL Championship on Reddit"},
			{Name: "Football - The People's Sport", URL: "https://www.reddit.com/r/football/.rss?format=xml", Description: "Home of Football. News, Rumours, Analysis, gossip and much more."},
			{Name: "Football News, Live Scores, Results &amp; Transfers | Goal.com", URL: "https://www.goal.com/feeds/en/news", Description: "The latest football news, rumours, transfers and match reports from around th..."},
			{Name: "Football365", URL: "https://www.football365.com/feed", Description: "Views, Live Matches, Gossip & more | Football365.com"},
			{Name: "Soccer News", URL: "https://www.soccernews.com/feed", Description: "The Latest Soccer News from around the globe."},
		},
	},
	{
		Name: "Funny",
		Feeds: []Feed{
			{Name: "AwkwardFamilyPhotos.com", URL: "https://awkwardfamilyphotos.com/feed/", Description: "Spreading the Awkwardness"},
			{Name: "Cracked: All Posts", URL: "http://feeds.feedburner.com/CrackedRSS", Description: "Every post that goes up on Cracked.com"},
			{Name: "Explosm.net", URL: "http://feeds.feedburner.com/Explosm", Description: "Flash Animations, Daily Comics and more!"},
			{Name: "FAIL Blog", URL: "http://feeds.feedburner.com/failblog", Description: "The internet has generated a huge amount of laughs from cats and FAILS. And w..."},
			{Name: "I Can Has Cheezburger?", URL: "http://feeds.feedburner.com/icanhascheezburger", Description: "World's largest collection of cat memes and other animals"},
			{Name: "PHD Comics", URL: "http://phdcomics.com/gradfeed.php", Description: "Providing global up-to-the-minute procrastination!"},
			{Name: "Penny Arcade", URL: "https://www.penny-arcade.com/feed", Description: "News Fucker 6000"},
			{Name: "PostSecret", URL: "https://postsecret.com/feed/?alt=rss", Description: "PostSecret"},
			{Name: "Saturday Morning Breakfast Cereal", URL: "https://www.smbc-comics.com/comic/rss", Description: "Latest Saturday Morning Breakfast Cereal comics and news"},
			{Name: "The Bloggess", URL: "https://thebloggess.com/feed/", Description: "Like Mother Teresa, only better."},
			{Name: "The Daily WTF", URL: "http://syndication.thedailywtf.com/TheDailyWtf", Description: "Curious Perversions in Information Technology"},
			{Name: "The Oatmeal - Comics by Matthew Inman", URL: "http://feeds.feedburner.com/oatmealfeed", Description: "I make comics about science, cats, technology, and sometimes goats."},
			{Name: "The Onion", URL: "https://www.theonion.com/rss", Description: "America's Finest News Source."},
			{Name: "xkcd.com", URL: "https://xkcd.com/rss.xml", Description: "xkcd.com: A webcomic of romance and math humor."},
		},
	},
	{
		Name: "Gaming",
		Feeds: []Feed{
			{Name: "Escapist Magazine", URL: "https://www.escapistmagazine.com/v2/feed/", Description: "Everything fun"},
			{Name: "Eurogamer.net", URL: "https://www.eurogamer.net/?format=rss", Description: "Bad puns and video games since 1999."},
			{Name: "Gamasutra News", URL: "http://feeds.feedburner.com/GamasutraNews", Description: "Gamasutra News"},
			{Name: "GameSpot - All Content", URL: "https://www.gamespot.com/feeds/mashup/", Description: "GameSpot's Everything Feed! All the latest from GameSpot"},
			{Name: "IGN All", URL: "http://feeds.ign.com/ign/all", Description: "The latest IGN news, reviews and videos about video games, movies, TV, tech a..."},
			{Name: "Indie Games Plus", URL: "https://indiegamesplus.com/feed", Description: "Creative, Personal, Passionate Digital Experiences"},
			{Name: "Kotaku", URL: "https://kotaku.com/rss", Description: "Gaming Reviews, News, Tips and More."},
			{Name: "Makeup and Beauty Blog | Makeup Reviews, Swatches and How-To Makeup", URL: "https://www.makeupandbeautyblog.com/feed/", Description: "A beauty blog blooming with fresh makeup reviews, swatches and beauty tips fr..."},
			{Name: "PlayStation.Blog", URL: "http://feeds.feedburner.com/psblog", Description: "Official PlayStation Blog for news and video updates on PlayStation, PS5, PS4..."},
			{Name: "Polygon -  All", URL: "https://www.polygon.com/rss/index.xml", Description: "Polygon -  All"},
			{Name: "Rock, Paper, Shotgun", URL: "http://feeds.feedburner.com/RockPaperShotgun", Description: "This is a feed of all the latest articles from Rock Paper Shotgun."},
			{Name: "Steam RSS News Feed", URL: "https://store.steampowered.com/feeds/news.xml", Description: "All Steam news, all the time!"},
			{Name: "The Ancient Gaming Noob", URL: "http://feeds.feedburner.com/TheAncientGamingNoob", Description: "Veni, Vidi, Scripsi"},
			{Name: "TouchArcade - iPhone, iPad, Android Games Forum", URL: "https://toucharcade.com/community/forums/-/index.rss", Description: "iPhone and iPad Games and News"},
			{Name: "Xbox's Major Nelson", URL: "https://majornelson.com/feed/", Description: "Xbox news & facts direct from the source"},
			{Name: "r/gaming", URL: "https://www.reddit.com/r/gaming.rss", Description: "A subreddit for (almost) anything related to games - video games, board games..."},
		},
	},
	{
		Name: "History",
		Feeds: []Feed{
			{Name: "30 For 30 Podcasts", URL: "https://feeds.megaphone.fm/ESP5765452710", Description: "Original audio documentaries from the makers of the acclaimed 30 for 30 film ..."},
			{Name: "Blog Feed", URL: "https://americanhistory.si.edu/blog/feed", Description: "Blog Feed"},
			{Name: "Dan Carlin's Hardcore History", URL: "https://feeds.feedburner.com/dancarlin/history?format=xml", Description: "In"},
			{Name: "History in 28-minutes", URL: "https://www.historyisnowmagazine.com/blog?format=RSS", Description: "We create a variety of exclusive modern international and American history co..."},
			{Name: "HistoryNet", URL: "http://www.historynet.com/feed", Description: "HistoryNet.com contains daily features, photo galleries and over 5,000 articl..."},
			{Name: "Lore", URL: "https://feeds.megaphone.fm/lore", Description: "Lore is a bi-weekly podcast (as well as a TV show and book series) about dark..."},
			{Name: "Revisionist History", URL: "https://feeds.megaphone.fm/revisionisthistory", Description: "Revisionist History is Malcolm Gladwell's journey through the overlooked and ..."},
			{Name: "The History Reader", URL: "https://www.thehistoryreader.com/feed/", Description: "A History Blog from St. Martin’s Press"},
			{Name: "Throughline", URL: "https://feeds.npr.org/510333/podcast.xml", Description: "The past is never past. Every headline has a history. Join us every week as w..."},
			{Name: "You Must Remember This", URL: "https://feeds.megaphone.fm/YMRT7068253588", Description: "You Must Remember This is a storytelling podcast exploring the secret and/or ..."},
			{Name: "the memory palace", URL: "http://feeds.thememorypalace.us/thememorypalace", Description: "<p>the memory palace</p>"},
		},
	},
	{
		Name: "Interior design",
		Feeds: []Feed{
			{Name: "Apartment Therapy| Saving the world, one room at a time", URL: "https://www.apartmenttherapy.com/design.rss", Description: "Saving the world, one room at a time"},
			{Name: "Better Living Through Design", URL: "http://www.betterlivingthroughdesign.com/feed/", Description: "Better Living Through Design ™ -- Your Design Guide to Home & Style."},
			{Name: "Blog - decor8", URL: "https://www.decor8blog.com/blog?format=rss", Description: "<p>An expertly curated guide for stylish home and living. Holly Becker helps ..."},
			{Name: "Core77", URL: "http://feeds.feedburner.com/core77/blog", Description: "Launched in 1995, Core77 serves a devoted global audience of design professio..."},
			{Name: "Design MilkInterior Design – Design Milk", URL: "https://design-milk.com/category/interior-design/feed/", Description: "Dedicated to modern design"},
			{Name: "Fubiz Media", URL: "http://feeds.feedburner.com/fubiz", Description: "The latest creative news from Fubiz about art, design and pop-culture."},
			{Name: "Ideal Home", URL: "https://www.idealhome.co.uk/feed", Description: "Ideal Home"},
			{Name: "In My Own Style", URL: "https://inmyownstyle.com/feed", Description: "Top Budget DIY Home Decorating Blog + Creative Lifestyle Ideas + Tutorials."},
			{Name: "Inhabitat - Green Design, Innovation, Architecture, Green Building", URL: "https://inhabitat.com/design/feed/", Description: "Green design & innovation for a better world"},
			{Name: "Interior Design (Interior Architecture)", URL: "https://www.reddit.com/r/InteriorDesign/.rss", Description: "Interior Design is the art and science of understanding people's behavior to ..."},
			{Name: "Interior Design Ideas", URL: "http://www.home-designing.com/feed", Description: "Inspirational Interior Design Ideas"},
			{Name: "Interior Design Latest", URL: "https://www.interiordesign.net/rss/", Description: "Latest content from interiordesign.net"},
			{Name: "Interiors – Dezeen", URL: "https://www.dezeen.com/interiors/feed/", Description: "architecture and design magazine"},
			{Name: "Liz Marie Blog", URL: "https://www.lizmarieblog.com/feed/", Description: "Liz Marie Blog"},
			{Name: "The Design Files | Australia's most popular design blog.The Design Files | Australia's most popular design blog.", URL: "https://thedesignfiles.net/feed/", Description: "Australia's most popular design blog."},
			{Name: "The Inspired Room", URL: "https://theinspiredroom.net/feed/", Description: "Voted Readers' Favorite Top Decorating Blog Better Homes and Gardens, Decorat..."},
			{Name: "Thrifty Decor Chick", URL: "http://feeds.feedburner.com/blogspot/ZBcZ", Description: "Thrifty decorating blog with easy DIY, decor, entertaining and organizing ide..."},
			{Name: "Trendir", URL: "https://www.trendir.com/feed/", Description: "The Latest Trends in Modern House Design and Decorating"},
			{Name: "Yanko Design", URL: "http://feeds.feedburner.com/yankodesign", Description: "Modern Industrial Design News"},
			{Name: "Yatzer RSS Feed", URL: "https://www.yatzer.com/rss.xml", Description: "Yatzer RSS Feed"},
			{Name: "Young House Love", URL: "https://www.younghouselove.com/feed/", Description: "DIY Home Decorating Projects, Tutorials, & Shenanigans"},
			{Name: "decoist", URL: "https://www.decoist.com/feed/", Description: "Obsessed with the cult of beauty"},
			{Name: "design – designboom | architecture & design magazine", URL: "https://www.designboom.com/design/feed", Description: "designboom magazine | your first source for architecture, design & art news"},
			{Name: "sfgirlbybay", URL: "https://www.sfgirlbybay.com/feed/", Description: "bohemian modern style from a san francisco girl."},
		},
	},
	{
		Name: "Movies",
		Feeds: []Feed{
			{Name: "/Film", URL: "https://feeds2.feedburner.com/slashfilm", Description: "Movie News and Reviews This feed is for non commercial use. Content Copyright..."},
			{Name: "Ain't It Cool News Feed", URL: "https://www.aintitcool.com/node/feed/", Description: "The best in movie, TV, DVD and comic book news."},
			{Name: "ComingSoon.net", URL: "https://www.comingsoon.net/feed", Description: "New Movies, Movie Trailers, TV, Streaming, Anime & Video Game News"},
			{Name: "Deadline", URL: "https://deadline.com/feed/", Description: "Hollywood Entertainment Breaking News"},
			{Name: "Film School Rejects", URL: "https://filmschoolrejects.com/feed/", Description: "Movies, TV, and Culture"},
			{Name: "FirstShowing.net", URL: "https://www.firstshowing.net/feed/", Description: "Connecting Hollywood with its Audience"},
			{Name: "IndieWire", URL: "https://www.indiewire.com/feed", Description: "The Voice of Creative Independence"},
			{Name: "Movie News and Discussion", URL: "https://reddit.com/r/movies/.rss", Description: "News & Discussion about Major Motion Pictures"},
			{Name: "Movies", URL: "https://www.bleedingcool.com/movies/feed/", Description: "Comics, Movies, TV, Games"},
			{Name: "The A.V. Club", URL: "https://film.avclub.com/rss", Description: "Pop culture obsessives writing for the pop culture obsessed."},
			{Name: "Variety", URL: "https://variety.com/feed/", Description: "Variety"},
		},
	},
	{
		Name: "Music",
		Feeds: []Feed{
			{Name: "Billboard", URL: "https://www.billboard.com/articles/rss.xml", Description: "Just another WordPress site"},
			{Name: "Consequence", URL: "http://consequenceofsound.net/feed", Description: "Music, Film, TV and Pop Culture News for the Mainstream and Underground"},
			{Name: "EDM.com - The Latest Electronic Dance Music News, Reviews & Artists", URL: "https://edm.com/.rss/full/", Description: "The world's foremost authority on EDM : free music downloads, artist intervie..."},
			{Name: "Metal Injection", URL: "http://feeds.feedburner.com/metalinjection", Description: "Heavy metal news, metal music videos, tour dates, live footage, exclusive doc..."},
			{Name: "Music Business Worldwide", URL: "https://www.musicbusinessworldwide.com/feed/", Description: "News, jobs and analysis for the global music industry"},
			{Name: "RSS: News", URL: "http://pitchfork.com/rss/news", Description: "News content RSS feed"},
			{Name: "Song Exploder", URL: "http://songexploder.net/feed", Description: "A podcast where musicians take apart their songs, and piece by piece, tell th..."},
			{Name: "Your EDM", URL: "https://www.youredm.com/feed", Description: "EDM News & Electronic Dance Music Reviews"},
		},
	},
	{
		Name: "News",
		Feeds: []Feed{
			{Name: "BBC News - World", URL: "http://feeds.bbci.co.uk/news/world/rss.xml", Description: "BBC News - World"},
			{Name: "CNN.com - RSS Channel - World", URL: "http://rss.cnn.com/rss/edition_world.rss", Description: "CNN.com delivers up-to-the-minute news and information on the latest top stor..."},
			{Name: "International: Top News And Analysis", URL: "https://www.cnbc.com/id/100727362/device/rss/rss.html", Description: "CNBC International is the world leader for news on business, technology, Chin..."},
			{Name: "NDTV News - World-news", URL: "http://feeds.feedburner.com/ndtvnews-world-news", Description: "NDTV.com provides the latest information from and in-depth coverage of India ..."},
			{Name: "NYT > World News", URL: "https://rss.nytimes.com/services/xml/rss/nyt/World.xml", Description: "NYT > World News"},
			{Name: "Top stories - Google News", URL: "https://news.google.com/rss", Description: "Google News"},
			{Name: "World", URL: "http://feeds.washingtonpost.com/rss/world", Description: "The Washington Post World section provides information and analysis of breaki..."},
			{Name: "World News", URL: "https://www.reddit.com/r/worldnews/.rss", Description: "A place for major news from around the world, excluding US-internal news."},
			{Name: "World News Headlines, Latest International News, World Breaking News - Times of India", URL: "https://timesofindia.indiatimes.com/rssfeeds/296589292.cms", Description: "World News: TOI brings the latest world news headlines, Current International..."},
			{Name: "World news | The Guardian", URL: "https://www.theguardian.com/world/rss", Description: "Latest World news news, comment and analysis from the Guardian, the world's l..."},
			{Name: "Yahoo News - Latest News & Headlines", URL: "https://www.yahoo.com/news/rss", Description: "The latest news and headlines from Yahoo! News. Get breaking news stories and..."},
		},
	},
	{
		Name: "Personal finance",
		Feeds: []Feed{
			{Name: "Afford Anything", URL: "https://affordanything.com/feed/", Description: "You Can Afford Anything ... Just Not Everything. What's It Gonna Be?"},
			{Name: "Blog – Student Loan Hero", URL: "https://studentloanhero.com/blog/feed", Description: "Blog – Student Loan Hero"},
			{Name: "Budgets Are Sexy", URL: "https://feeds2.feedburner.com/budgetsaresexy", Description: "Budgets Are Sexy"},
			{Name: "Financial Samurai", URL: "https://www.financialsamurai.com/feed/", Description: "Slicing Through Money's Mysteries"},
			{Name: "Frugalwoods", URL: "https://feeds.feedburner.com/Frugalwoods", Description: "Financial independence and simple living"},
			{Name: "Get Rich Slowly", URL: "https://www.getrichslowly.org/feed/", Description: "personal finance that makes cents"},
			{Name: "Good Financial Cents®", URL: "https://www.goodfinancialcents.com/feed/", Description: "Making Cents Of Investing and Financial Planning"},
			{Name: "I Will Teach You To Be Rich", URL: "https://www.iwillteachyoutoberich.com/feed/", Description: "by Ramit Sethi"},
			{Name: "Learn To Trade The Market", URL: "https://www.learntotradethemarket.com/feed", Description: "Learn Price Action Trading with Nial Fuller"},
			{Name: "Making Sense Of Cents", URL: "https://www.makingsenseofcents.com/feed", Description: "Learn how to make extra money, how to save money, how to start a blog, and more."},
			{Name: "Millennial Money", URL: "https://millennialmoney.com/feed/", Description: "Next Level Personal Finance"},
			{Name: "MintLife Blog", URL: "https://blog.mint.com/feed/", Description: "Personal Finance News & Advice"},
			{Name: "Money Crashers", URL: "https://www.moneycrashers.com/feed/", Description: "Personal Finance Guide to Turn the Tables on Money"},
			{Name: "Money Saving Mom®", URL: "https://moneysavingmom.com/feed/", Description: "Saving Families Money Since 2007"},
			{Name: "Money Under 30", URL: "https://www.moneyunder30.com/feed", Description: "Just another Word Press site"},
			{Name: "MoneyNing", URL: "http://feeds.feedburner.com/MoneyNing", Description: "Sharing insights since 2007 on carefully saving money, investing, frugal livi..."},
			{Name: "MyWifeQuitHerJob.com", URL: "https://mywifequitherjob.com/feed/", Description: "Starting An Online Store So Your Spouse Can Quit And Stay At Home With The Kids"},
			{Name: "Nerd's Eye View | Kitces.com", URL: "http://feeds.feedblitz.com/kitcesnerdseyeview&x=1", Description: "Commentary from Michael Kitces on Financial Planning News & Strategies"},
			{Name: "NerdWallet", URL: "https://www.nerdwallet.com/blog/feed/", Description: "NerdWallet is a free tool to find you the best credit cards, cd rates, saving..."},
			{Name: "Oblivious Investor", URL: "https://obliviousinvestor.com/feed/", Description: "Low-Maintenance Investing with Index Funds and ETFs"},
			{Name: "Personal Finance", URL: "https://reddit.com/r/personalfinance/.rss", Description: "Learn about budgeting, saving, getting out of debt, credit, investing, and re..."},
			{Name: "SavingAdvice.com Blog", URL: "https://www.savingadvice.com/feed/", Description: "Bridging the gap between saving money and investing"},
			{Name: "Side Hustle Nation", URL: "https://www.sidehustlenation.com/feed", Description: "Amplify your earning power"},
			{Name: "The College Investor", URL: "https://thecollegeinvestor.com/feed/", Description: "Student Loans, Investing, Building Wealth"},
			{Name: "The Dough Roller", URL: "https://www.doughroller.net/feed/", Description: "Money Management and Personal Finance | The Dough Roller"},
			{Name: "The Penny Hoarder", URL: "https://www.thepennyhoarder.com/feed/", Description: "The Penny Hoarder is one of the largest personal finance websites in America,..."},
			{Name: "Well Kept Wallet", URL: "https://wellkeptwallet.com/feed/", Description: "Well Kept Wallet"},
			{Name: "Wise Bread", URL: "http://feeds.killeraces.com/wisebread", Description: "Living large on a small budget"},
		},
	},
	{
		Name: "Photography",
		Feeds: []Feed{
			{Name: "500px", URL: "https://iso.500px.com/feed/", Description: "500px"},
			{Name: "500px:", URL: "https://500px.com/editors.rss", Description: "on 500px."},
			{Name: "Big Picture", URL: "https://www.bostonglobe.com/rss/bigpicture", Description: "News Stories in Photographs from the Boston Globe."},
			{Name: "Canon Rumors – Your best source for Canon rumors, leaks and gossip", URL: "https://www.canonrumors.com/feed/", Description: "Canon Rumors – Your best source for Canon rumors, leaks and gossip"},
			{Name: "Digital Photography School", URL: "https://feeds.feedburner.com/DigitalPhotographySchool", Description: "Digital Photography Tips and Tutorials"},
			{Name: "Light Stalking", URL: "https://www.lightstalking.com/feed/", Description: "Illuminating Your Passion"},
			{Name: "Lightroom Killer Tips", URL: "https://lightroomkillertips.com/feed/", Description: "Lightroom Presets, Videos, Tips and News"},
			{Name: "One Big Photo", URL: "http://feeds.feedburner.com/OneBigPhoto", Description: "a Picture is Worth a Thousand Words | Submit Your Photo | World's Best Photog..."},
			{Name: "PetaPixel", URL: "https://petapixel.com/feed/", Description: "Photography and Camera News, Reviews, and Inspiration"},
			{Name: "Strobist", URL: "http://feeds.feedburner.com/blogspot/WOBq", Description: "Learn How to Light."},
			{Name: "Stuck in Customs", URL: "https://stuckincustoms.com/feed/", Description: "Trey Ratcliff's Travel Photography blog with daily inspiration to motivate you!"},
			{Name: "The Sartorialist", URL: "https://feeds.feedburner.com/TheSartorialist", Description: "Just another WordPress site"},
		},
	},
	{
		Name: "Programming",
		Feeds: []Feed{
			{Name: "Better Programming - Medium", URL: "https://medium.com/feed/better-programming", Description: "Advice for programmers. - Medium"},
			{Name: "Code as Craft", URL: "https://codeascraft.com/feed/atom/", Description: "The Engineering Blog from Etsy"},
			{Name: "CodeNewbie", URL: "http://feeds.codenewbie.org/cnpodcast.xml", Description: "Stories and interviews from people on their coding journey."},
			{Name: "Coding Horror", URL: "https://feeds.feedburner.com/codinghorror", Description: "programming and human factors"},
			{Name: "Complete Developer Podcast", URL: "https://completedeveloperpodcast.com/feed/podcast/", Description: "A podcast by coders for coders about all aspects of life as a developer."},
			{Name: "Dan Abramov's Overreacted Blog RSS Feed", URL: "https://overreacted.io/rss.xml", Description: "Personal blog by Dan Abramov. I explain with words and code."},
			{Name: "Developer Tea", URL: "https://feeds.simplecast.com/dLRotFGk", Description: "Developer Tea exists to help driven developers connect to their ultimate purp..."},
			{Name: "English (US)", URL: "https://blog.twitter.com/engineering/en_us/blog.rss", Description: "Information from Twitter's engineering team about our tools, technology and s..."},
			{Name: "FLOSS Weekly (Audio)", URL: "https://feeds.twit.tv/floss.xml", Description: "We're not talking dentistry here; FLOSS all about Free Libre Open Source Soft..."},
			{Name: "Facebook Engineering", URL: "https://engineering.fb.com/feed/", Description: "Facebook Engineering Blog"},
			{Name: "GitLab", URL: "https://about.gitlab.com/atom.xml", Description: "GitLab"},
			{Name: "Google Developers Blog", URL: "http://feeds.feedburner.com/GDBcode", Description: "Blog of our latest news, updates, and stories for developers"},
			{Name: "Google TechTalks", URL: "https://www.youtube.com/feeds/videos.xml?user=GoogleTechTalks", Description: "Google TechTalks"},
			{Name: "HackerNoon.com - Medium", URL: "https://medium.com/feed/hackernoon", Description: "Elijah McClain, George Floyd, Eric Garner, Breonna Taylor, Ahmaud Arbery, Mic..."},
			{Name: "Hanselminutes with Scott Hanselman", URL: "https://feeds.simplecast.com/gvtxUiIf", Description: "Hanselminutes is Fresh Air for Developers. A weekly commute-time podcast that..."},
			{Name: "InfoQ", URL: "https://feed.infoq.com", Description: "InfoQ feed"},
			{Name: "Instagram Engineering - Medium", URL: "https://instagram-engineering.com/feed/", Description: "Stories from the people who build @Instagram - Medium"},
			{Name: "Java, SQL and jOOQ.", URL: "https://blog.jooq.org/feed", Description: "Best Practices and Lessons Learned from Writing Awesome Java and SQL Code. Ge..."},
			{Name: "JetBrains Blog", URL: "https://blog.jetbrains.com/feed", Description: "Developer Tools for Professionals and Teams"},
			{Name: "Joel on Software", URL: "https://www.joelonsoftware.com/feed/", Description: "Joel on Software"},
			{Name: "LinkedIn Engineering", URL: "https://engineering.linkedin.com/blog.rss.html", Description: "The official blog of the Engineering team at LinkedIn"},
			{Name: "Martin Fowler", URL: "https://martinfowler.com/feed.atom", Description: "Master feed of news and updates from martinfowler.com"},
			{Name: "Netflix TechBlog - Medium", URL: "https://netflixtechblog.com/feed", Description: "Learn about Netflix’s world class engineering efforts, company culture, produ..."},
			{Name: "Overflow - Buffer Resources", URL: "https://buffer.com/resources/overflow/rss/", Description: "In-depth ideas and guides to social media & online marketing strategy, publis..."},
			{Name: "Podcast – Software Engineering Daily", URL: "https://softwareengineeringdaily.com/category/podcast/feed", Description: "Podcast – Software Engineering Daily"},
			{Name: "Posts on &> /dev/null", URL: "https://www.thirtythreeforty.net/posts/index.xml", Description: "Recent content in Posts on &> /dev/null"},
			{Name: "Prezi Engineering - Medium", URL: "https://engineering.prezi.com/feed", Description: "The things we learn as we build our products - Medium"},
			{Name: "Programming Throwdown", URL: "http://feeds.feedburner.com/ProgrammingThrowdown", Description: "Programming Throwdown educates Computer Scientists and Software Engineers on ..."},
			{Name: "Programming – The Crazy Programmer", URL: "https://www.thecrazyprogrammer.com/category/programming/feed", Description: "Programming, Design and Development"},
			{Name: "Robert Heaton | Blog", URL: "https://robertheaton.com/feed.xml", Description: "Software engineer. One-track lover down a two-way lane"},
			{Name: "Scott Hanselman's Blog", URL: "http://feeds.hanselman.com/ScottHanselman", Description: "Scott Hanselman on Programming, User Experience, The Zen of Computers and Lif..."},
			{Name: "Scripting News", URL: "http://scripting.com/rss.xml", Description: "It's even worse than it appears."},
			{Name: "Signal v. Noise", URL: "https://m.signalvnoise.com/feed/", Description: "Strong opinions and shared thoughts on design, business, and tech. By the mak..."},
			{Name: "Slack Engineering", URL: "https://slack.engineering/feed", Description: "Slack Engineering"},
			{Name: "Software Defined Talk", URL: "https://feeds.fireside.fm/sdt/rss", Description: "A weekly podcast covering all the news and events in Enterprise Software and ..."},
			{Name: "Software Engineering Radio - The Podcast for Professional Software Developers", URL: "http://feeds.feedburner.com/se-radio", Description: "Software Engineering Radio is a podcast targeted at the professional software..."},
			{Name: "SoundCloud Backstage Blog", URL: "https://developers.soundcloud.com/blog/blog.rss", Description: "SoundCloud's developer blog."},
			{Name: "Spotify Engineering", URL: "https://labs.spotify.com/feed/", Description: "Spotify’s official technology blog"},
			{Name: "Stack Abuse", URL: "https://stackabuse.com/rss/", Description: "Learn Python, Java, JavaScript/Node, Machine Learning, and Web Development th..."},
			{Name: "Stack Overflow Blog", URL: "https://stackoverflow.blog/feed/", Description: "Essays, opinions, and advice on the act of computer programming from Stack Ov..."},
			{Name: "The 6 Figure Developer", URL: "http://6figuredev.com/feed/rss/", Description: "Helping others reach their potential"},
			{Name: "The Airbnb Tech Blog - Medium", URL: "https://medium.com/feed/airbnb-engineering", Description: "Creative engineers and data scientists building a world where you can belong ..."},
			{Name: "The GitHub Blog", URL: "https://github.blog/feed/", Description: "Updates, ideas, and inspiration from GitHub to help developers build and desi..."},
			{Name: "The PIT Show: Reflections and Interviews in the Tech World", URL: "https://feeds.transistor.fm/productivity-in-tech-podcast", Description: "This is the show where I sit down with people in the tech space and talk abou..."},
			{Name: "The Rabbit Hole: The Definitive Developer's Podcast", URL: "http://therabbithole.libsyn.com/rss", Description: "Welcome to The Rabbit Hole, the definitive developers podcast. If you are a s..."},
			{Name: "The Stack Overflow Podcast", URL: "https://feeds.simplecast.com/XA_851k3", Description: "The Stack Overflow podcast is a weekly conversation about working in software..."},
			{Name: "The Standup", URL: "https://feeds.fireside.fm/standup/rss", Description: "A podcast that delves into the obstacles and successes involved in creating, ..."},
			{Name: "The Women in Tech Show: A Technical Podcast", URL: "https://thewomenintechshow.com/category/podcast/feed/", Description: "A podcast about what we work on, not what it feels like to be a woman in tech..."},
			{Name: "programming", URL: "https://www.reddit.com/r/programming/.rss", Description: "Computer Programming"},
		},
	},
	{
		Name: "Science",
		Feeds: []Feed{
			{Name: "60-Second Science", URL: "http://rss.sciam.com/sciam/60secsciencepodcast", Description: "Tune in every weekday for quick reports and commentaries on the world of scie..."},
			{Name: "BBC News - Science & Environment", URL: "http://feeds.bbci.co.uk/news/science_and_environment/rss.xml", Description: "BBC News - Science & Environment"},
			{Name: "Discovery", URL: "https://podcasts.files.bbci.co.uk/p002w557.rss", Description: "Explorations in the world of science."},
			{Name: "FlowingData", URL: "https://flowingdata.com/feed", Description: "Strength in Numbers"},
			{Name: "Gizmodo", URL: "https://gizmodo.com/tag/science/rss", Description: "We come from the future"},
			{Name: "Hidden Brain", URL: "https://feeds.npr.org/510308/podcast.xml", Description: "Shankar Vedantam uses science and storytelling to reveal the unconscious patt..."},
			{Name: "Invisibilia", URL: "https://feeds.npr.org/510307/podcast.xml", Description: "Unseeable forces control human behavior and shape our ideas, beliefs, and ass..."},
			{Name: "Latest Science News -- ScienceDaily", URL: "https://www.sciencedaily.com/rss/all.xml", Description: "Latest Science News -- ScienceDaily"},
			{Name: "NYT > Science", URL: "https://rss.nytimes.com/services/xml/rss/nyt/Science.xml", Description: "NYT > Science"},
			{Name: "Nature", URL: "https://www.nature.com/nature.rss", Description: "Nature is the foremost international weekly scientific journal in the world a..."},
			{Name: "Phys.org - latest science and technology news stories", URL: "https://phys.org/rss-feed/", Description: "Phys.org internet news portal provides the latest news on science including: ..."},
			{Name: "Probably Science", URL: "https://probablyscience.libsyn.com/rss", Description: "Professional comedians with so-so STEM pedigrees take you through this week i..."},
			{Name: "Radiolab", URL: "http://feeds.wnyc.org/radiolab", Description: "Radiolab is one of the most beloved podcasts and public radio shows in the wo..."},
			{Name: "Reddit Science", URL: "https://reddit.com/r/science/.rss", Description: "This community is a place to share and discuss new scientific research. Read ..."},
			{Name: "Sawbones: A Marital Tour of Misguided Medicine", URL: "https://feeds.simplecast.com/y1LF_sn2", Description: "Join Dr. Sydnee McElroy and her husband Justin McElroy for a tour of all the ..."},
			{Name: "Science Latest", URL: "https://www.wired.com/feed/category/science/latest/rss", Description: "Channel Description"},
			{Name: "Science Vs", URL: "http://feeds.gimletmedia.com/ScienceVs", Description: "There are a lot of fads, blogs and strong opinions, but then there’s SCIENCE...."},
			{Name: "Science-Based Medicine", URL: "https://sciencebasedmedicine.org/feed/", Description: "Exploring issues and controversies in the relationship between science and me..."},
			{Name: "Scientific American Content: Global", URL: "http://rss.sciam.com/ScientificAmerican-Global", Description: "Science news and technology updates from Scientific American"},
			{Name: "TED Talks Daily (SD video)", URL: "https://pa.tedcdn.com/feeds/talks.rss", Description: "TED is a nonprofit devoted to ideas worth spreading. On this video feed, you'..."},
			{Name: "The Infinite Monkey Cage", URL: "https://podcasts.files.bbci.co.uk/b00snr0w.rss", Description: "Witty, irreverent look at the world through scientists' eyes. With Brian Cox ..."},
			{Name: "This Week in Science – The Kickass Science Podcast", URL: "http://www.twis.org/feed/", Description: "The kickass science and technology radio show that delivers an irreverent loo..."},
		},
	},
	{
		Name: "Space",
		Feeds: []Feed{
			{Name: "/r/space: news, articles and discussion", URL: "https://www.reddit.com/r/space/.rss?format=xml", Description: "Share & discuss informative content on: * Astrophysics * Cosmology * Space Ex..."},
			{Name: "NASA Breaking News", URL: "https://www.nasa.gov/rss/dyn/breaking_news.rss", Description: "A RSS news feed containing the latest NASA news articles and press releases."},
			{Name: "New Scientist - Space", URL: "https://www.newscientist.com/subject/space/feed/", Description: "New Scientist - Space"},
			{Name: "Sky & Telescope", URL: "https://www.skyandtelescope.com/feed/", Description: "The essential guide to astronomy"},
			{Name: "Space | The Guardian", URL: "https://www.theguardian.com/science/space/rss", Description: "Latest news and features from theguardian.com, the world's leading liberal voice"},
			{Name: "Space.com", URL: "https://www.space.com/feeds/all", Description: "Get the latest space exploration, innovation and astronomy news. Space.com ce..."},
			{Name: "SpaceX", URL: "https://www.youtube.com/feeds/videos.xml?user=spacexchannel", Description: "SpaceX"},
		},
	},
	{
		Name: "Sports",
		Feeds: []Feed{
			{Name: "BBC Sport - Sport", URL: "http://feeds.bbci.co.uk/sport/rss.xml", Description: "BBC Sport - Sport"},
			{Name: "Reddit Sports", URL: "https://www.reddit.com/r/sports.rss", Description: "Sports News and Highlights from the NFL, NBA, NHL, MLB, MLS, and leagues arou..."},
			{Name: "Sports News - Latest Sports and Football News | Sky News", URL: "http://feeds.skynews.com/feeds/rss/sports.xml", Description: "The best sports coverage from around the world, covering: Football, Cricket, ..."},
			{Name: "Sportskeeda", URL: "https://www.sportskeeda.com/feed", Description: "Sports Writers Unite"},
			{Name: "Yahoo! Sports - News, Scores, Standings, Rumors, Fantasy Games", URL: "https://sports.yahoo.com/rss/", Description: "Yahoo! Sports - Comprehensive news, scores, standings, fantasy games, rumors,..."},
			{Name: "www.espn.com - TOP", URL: "https://www.espn.com/espn/rss/news", Description: "Latest TOP news from www.espn.com"},
		},
	},
	{
		Name: "Startups",
		Feeds: []Feed{
			{Name: "AVC", URL: "https://avc.com/feed/", Description: "Musings of a VC in NYC"},
			{Name: "Both Sides of the Table - Medium", URL: "https://bothsidesofthetable.com/feed", Description: "Perspectives of a 2x entrepreneur turned VC at @UpfrontVC, the largest and mo..."},
			{Name: "Entrepreneur", URL: "http://feeds.feedburner.com/entrepreneur/latest", Description: "The latest stories from Entrepreneur."},
			{Name: "Feld Thoughts", URL: "https://feld.com/feed", Description: "Feld Thoughts"},
			{Name: "Forbes - Entrepreneurs", URL: "https://www.forbes.com/entrepreneurs/feed/", Description: "Forbes - Entrepreneurs"},
			{Name: "GaryVee", URL: "https://www.youtube.com/feeds/videos.xml?user=GaryVaynerchuk", Description: "GaryVee"},
			{Name: "Hacker News: Front Page", URL: "https://hnrss.org/frontpage", Description: "Hacker News RSS"},
			{Name: "Inc.com", URL: "https://www.inc.com/rss/", Description: "Inc.com, the daily resource for entrepreneurs."},
			{Name: "Inside Intercom", URL: "https://www.intercom.com/blog/feed", Description: "Product, Marketing, and Customer Support Blog"},
			{Name: "Marie Forleo", URL: "https://www.youtube.com/feeds/videos.xml?user=marieforleo", Description: "Marie Forleo"},
			{Name: "Masters of Scale with Reid Hoffman", URL: "https://rss.art19.com/masters-of-scale", Description: "<p>The best startup advice from Silicon Valley and beyond. Iconic CEOs — from..."},
			{Name: "Paul Graham: Essays", URL: "http://www.aaronsw.com/2002/feeds/pgessays.rss", Description: "Scraped feed provided by aaronsw.com"},
			{Name: "Product Hunt — The best new products, every day", URL: "https://www.producthunt.com/feed", Description: "Product Hunt — The best new products, every day"},
			{Name: "Quick Sprout", URL: "https://www.quicksprout.com/rss", Description: "Start and Grow Your Business"},
			{Name: "Small Business Trends", URL: "https://feeds2.feedburner.com/SmallBusinessTrends", Description: "Founded in 2003, Small Business Trends is an award-winning online publication..."},
			{Name: "Smart Passive Income", URL: "http://feeds.feedburner.com/smartpassiveincome", Description: "Smart Passive Income"},
			{Name: "Springwise", URL: "https://www.springwise.com/feed", Description: "Innovation That Matters"},
			{Name: "Steve Blank", URL: "https://steveblank.com/feed/", Description: "Innovation and Entrepreneurship"},
			{Name: "The Startup Junkies Podcast", URL: "https://startupjunkie.libsyn.com/rss", Description: "The Startup Junkies podcast is hosted by Jeff Amerine and his team at Startup..."},
			{Name: "The Tim Ferriss Show", URL: "https://rss.art19.com/tim-ferriss-show", Description: "Tim Ferriss is a self-experimenter and bestselling author, best known for The..."},
			{Name: "This Week in Startups - Video", URL: "http://feeds.feedburner.com/twistvid", Description: "Angel investor Jason Calacanis (Uber, Calm, Robinhood) interviews the world’s..."},
			{Name: "VentureBeat", URL: "https://feeds.feedburner.com/venturebeat/SZYF", Description: "Transformative tech coverage that matters"},
			{Name: "blog – Feld Thoughts", URL: "https://feld.com/archives/tag/blog/feed", Description: "blog – Feld Thoughts"},
		},
	},
	{
		Name: "Tech",
		Feeds: []Feed{
			{Name: "Accidental Tech Podcast", URL: "https://atp.fm/rss", Description: "Three nerds discussing tech, Apple, programming, and loosely related matters."},
			{Name: "Analog(ue)", URL: "https://www.relay.fm/analogue/feed", Description: "So many podcasts are about our digital devices. Analog(ue) is a show about ho..."},
			{Name: "Ars Technica", URL: "http://feeds.arstechnica.com/arstechnica/index", Description: "Serving the Technologist for more than a decade. IT news, reviews, and analysis."},
			{Name: "CNET", URL: "https://www.youtube.com/feeds/videos.xml?user=CNETTV", Description: "CNET"},
			{Name: "CNET News", URL: "https://www.cnet.com/rss/news/", Description: "CNET news editors and reporters provide top technology news, with investigati..."},
			{Name: "Clockwise", URL: "https://www.relay.fm/clockwise/feed", Description: "Clockwise is a rapid-fire discussion of current technology issues hosted by D..."},
			{Name: "Gizmodo", URL: "https://gizmodo.com/rss", Description: "We come from the future"},
			{Name: "Hacker News", URL: "https://news.ycombinator.com/rss", Description: "Links for the intellectually curious, ranked by readers."},
			{Name: "Lifehacker", URL: "https://lifehacker.com/rss", Description: "Do everything better"},
			{Name: "Linus Tech Tips", URL: "https://www.youtube.com/feeds/videos.xml?user=LinusTechTips", Description: "Linus Tech Tips"},
			{Name: "Marques Brownlee", URL: "https://www.youtube.com/feeds/videos.xml?user=marquesbrownlee", Description: "Marques Brownlee"},
			{Name: "Mashable", URL: "http://feeds.mashable.com/Mashable", Description: "Mashable is a leading source for news, information & resources for the Connec..."},
			{Name: "ReadWrite", URL: "https://readwrite.com/feed/", Description: "The Blog of Things"},
			{Name: "Reply All", URL: "https://feeds.megaphone.fm/replyall", Description: "Reply All"},
			{Name: "Rocket", URL: "https://www.relay.fm/rocket/feed", Description: "Countdown to excitement! Every week Christina Warren, Brianna Wu and Simone d..."},
			{Name: "Slashdot", URL: "http://rss.slashdot.org/Slashdot/slashdotMain", Description: "News for nerds, stuff that matters"},
			{Name: "Stratechery by Ben Thompson", URL: "http://stratechery.com/feed/", Description: "On the business, strategy, and impact of technology."},
			{Name: "TechCrunch", URL: "http://feeds.feedburner.com/TechCrunch", Description: "TechCrunch is a group-edited blog that profiles the companies, products and e..."},
			{Name: "The Keyword", URL: "https://www.blog.google/rss/", Description: "The Keyword"},
			{Name: "The Next Web", URL: "https://thenextweb.com/feed/", Description: "Original and proudly opinionated perspectives for Generation T"},
			{Name: "The Verge", URL: "https://www.youtube.com/feeds/videos.xml?user=TheVerge", Description: "The Verge"},
			{Name: "The Verge -  All Posts", URL: "https://www.theverge.com/rss/index.xml", Description: "The Verge -  All Posts"},
			{Name: "The Vergecast", URL: "https://feeds.megaphone.fm/vergecast", Description: "Hello! This is The Vergecast, the flagship podcast of The Verge... and your l..."},
			{Name: "This Week in Tech (Audio)", URL: "https://feeds.twit.tv/twit.xml", Description: "Your first podcast of the week is the last word in tech. Join the top tech pu..."},
			{Name: "Unbox Therapy", URL: "https://www.youtube.com/feeds/videos.xml?user=unboxtherapy", Description: "Unbox Therapy"},
			{Name: "https://www.engadget.com/", URL: "https://www.engadget.com/rss.xml", Description: "https://www.engadget.com/"},
		},
	},
	{
		Name: "Television",
		Feeds: []Feed{
			{Name: "TV", URL: "https://www.bleedingcool.com/tv/feed/", Description: "Comics, Movies, TV, Games"},
			{Name: "TV Fanatic", URL: "https://www.tvfanatic.com/rss.xml", Description: "TV Fanatic"},
			{Name: "TVLine", URL: "https://tvline.com/feed/", Description: "TV News, Previews, Spoilers, Casting Scoop, Interviews"},
			{Name: "Television News and Discussion", URL: "https://reddit.com/r/television/.rss", Description: "‎"},
			{Name: "The A.V. Club", URL: "https://tv.avclub.com/rss", Description: "Pop culture obsessives writing for the pop culture obsessed."},
			{Name: "the TV addict", URL: "http://feeds.feedburner.com/thetvaddict/AXob", Description: "TV News, Previews, Spoilers and Reviews"},
		},
	},
	{
		Name: "Tennis",
		Feeds: []Feed{
			{Name: "BBC Sport - Tennis", URL: "http://feeds.bbci.co.uk/sport/tennis/rss.xml", Description: "BBC Sport - Tennis"},
			{Name: "Essential Tennis Podcast - Instruction, Lessons, Tips", URL: "https://feed.podbean.com/essentialtennis/feed.xml", Description: "Improve your tennis with the Essential Tennis Podcast, the very first podcast..."},
			{Name: "Grand Slam Fantasy Tennis", URL: "http://www.grandslamfantasytennis.com/feed/?x=1", Description: "Fantasy Tennis"},
			{Name: "Tennis - ATP World Tour", URL: "https://www.atptour.com/en/media/rss-feed/xml-feed", Description: "Headline News - powered by FeedBurner"},
			{Name: "Tennis News & Discussion", URL: "https://www.reddit.com/r/tennis/.rss", Description: "Tennis News & Discussion"},
			{Name: "peRFect Tennis", URL: "https://www.perfect-tennis.com/feed/", Description: "UK Tennis Blog"},
			{Name: "www.espn.com - TENNIS", URL: "https://www.espn.com/espn/rss/tennis/news", Description: "Latest TENNIS news from www.espn.com"},
		},
	},
	{
		Name: "Travel",
		Feeds: []Feed{
			{Name: "Atlas Obscura - Latest Articles and Places", URL: "https://www.atlasobscura.com/feeds/latest", Description: "New wonders and curiosities added to the Atlas."},
			{Name: "Live Life Travel", URL: "https://www.livelifetravel.world/feed/", Description: "Your travel partner for life"},
			{Name: "Lonely Planet Travel News", URL: "https://www.lonelyplanet.com/news/feed/atom/", Description: "Travel news and more from Lonely Planet"},
			{Name: "NYT > Travel", URL: "https://rss.nytimes.com/services/xml/rss/nyt/Travel.xml", Description: "NYT > Travel"},
			{Name: "Nomadic Matt's Travel Site", URL: "https://www.nomadicmatt.com/travel-blog/feed/", Description: "Travel Better, Cheaper, Longer"},
			{Name: "Travel | The Guardian", URL: "https://www.theguardian.com/uk/travel/rss", Description: "Latest travel news and reviews on UK and world holidays, travel guides to glo..."},
		},
	},
	{
		Name: "UI - UX",
		Feeds: []Feed{
			{Name: "Articles on Smashing Magazine — For Web Designers And Developers", URL: "https://www.smashingmagazine.com/feed", Description: "Recent content in Articles on Smashing Magazine — For Web Designers And Devel..."},
			{Name: "Boxes and Arrows", URL: "http://boxesandarrows.com/rss/", Description: "The design behind the design."},
			{Name: "Designer News Feed", URL: "https://www.designernews.co/?format=rss", Description: "All of the stories from the frontpage of Designer News"},
			{Name: "Inside Design", URL: "https://www.invisionapp.com/inside-design/feed", Description: "Thoughts on users, experience, and design"},
			{Name: "JUST™ Creative", URL: "https://feeds.feedburner.com/JustCreativeDesignBlog", Description: "Logo Designer, Graphic Designer, Graphic Design Portfolio, Logo Design, Logo,..."},
			{Name: "NN/g latest articles and announcements", URL: "https://www.nngroup.com/feed/rss/", Description: "The latest articles and announcements from Nielsen Norman Group"},
			{Name: "UX Blog – UX Studio", URL: "https://uxstudioteam.com/ux-blog/feed/", Description: "UX design blog about designing user experience for web and mobile apps with U..."},
			{Name: "UX Collective - Medium", URL: "https://uxdesign.cc/feed", Description: "We believe designers are thinkers as much as they are makers. Curated stories..."},
			{Name: "UX Movement", URL: "https://uxmovement.com/feed/", Description: "UX Movement"},
			{Name: "Usability Geek", URL: "https://usabilitygeek.com/feed/", Description: "Usability & User Experience (UX) Blog"},
			{Name: "User Experience", URL: "https://www.reddit.com/r/userexperience/.rss", Description: "User experience design is the process of enhancing user satisfaction by impro..."},
		},
	},
	{
		Name: "Web Development",
		Feeds: []Feed{
			{Name: "A List Apart: The Full Feed", URL: "https://alistapart.com/main/feed/", Description: "Articles for people who make web sites."},
			{Name: "CSS-Tricks", URL: "https://css-tricks.com/feed/", Description: "Tips, Tricks, and Techniques on using Cascading Style Sheets."},
			{Name: "Code Wall", URL: "https://www.codewall.co.uk/feed/", Description: "Web Development & Programming"},
			{Name: "David Walsh Blog", URL: "https://davidwalsh.name/feed", Description: "A blog featuring tutorials about JavaScript, HTML5, AJAX, PHP, CSS, WordPress..."},
			{Name: "Mozilla Hacks – the Web developer blog", URL: "https://hacks.mozilla.org/feed/", Description: "hacks.mozilla.org"},
			{Name: "Sink In - Tech and Travel", URL: "https://gosink.in/rss/", Description: "Sink In - Tech and Travel"},
			{Name: "Updates", URL: "https://developers.google.com/web/updates/rss.xml", Description: "The latest and freshest updates from the Web teams at Google. Chrome, V8, too..."},
		},
	},
	{
		Name: "iOS Development",
		Feeds: []Feed{
			{Name: "ALL SHOWS - Devchat.tv", URL: "https://feeds.feedwrench.com/all-shows-devchattv.rss", Description: "ALL SHOWS - Devchat.tv"},
			{Name: "Alberto De Bortoli", URL: "https://albertodebortoli.com/rss/", Description: "Principal Software Engineer @ Just Eat Takeaway. Based in London, made in Italy."},
			{Name: "Augmented Code", URL: "https://augmentedcode.io/feed/", Description: "Performant, sleek and elegant."},
			{Name: "Benoit Pasquier - Swift, Data and more", URL: "https://benoitpasquier.com/index.xml", Description: "Recent content on Benoit Pasquier - Swift, Data and more"},
			{Name: "Fabisevi.ch", URL: "https://www.fabisevi.ch/feed.xml", Description: "iOS Developer [@Twitter](https://twitter.com/mergesort)"},
			{Name: "Mobile A11y", URL: "https://mobilea11y.com/index.xml", Description: "Recent content on Mobile A11y"},
			{Name: "More Than Just Code podcast - iOS and Swift development, news and advice", URL: "https://feeds.fireside.fm/mtjc/rss", Description: "MTJC podcast is a show about mobile development. Each week Jaime Lopez, Mark ..."},
			{Name: "News - Apple Developer", URL: "https://developer.apple.com/news/rss/news.rss", Description: "News - Apple Developer"},
			{Name: "Ole Begemann", URL: "https://oleb.net/blog/atom.xml", Description: "Ole Begemann"},
			{Name: "Pavel Zak’s dev blog", URL: "https://nerdyak.tech/feed.xml", Description: "Yet another dev blog"},
			{Name: "Swift by Sundell", URL: "https://www.swiftbysundell.com/feed.rss", Description: "Weekly Swift articles, podcasts and tips by John Sundell"},
			{Name: "Swift by Sundell", URL: "https://swiftbysundell.com/feed.rss", Description: "Weekly Swift articles, podcasts and tips by John Sundell"},
			{Name: "SwiftRocks", URL: "https://swiftrocks.com/rss.xml", Description: "SwiftRocks is a blog about how Swift works and general iOS tips and tricks."},
			{Name: "The Atomic Birdhouse", URL: "https://atomicbird.com/index.xml", Description: "Recent content on The Atomic Birdhouse"},
			{Name: "Under the Radar", URL: "https://www.relay.fm/radar/feed", Description: "From development and design to marketing and support, Under the Radar is all ..."},
			{Name: "Use Your Loaf - iOS Development News & Tips", URL: "https://useyourloaf.com/blog/rss.xml", Description: "Recent content on Use Your Loaf - iOS Development News & Tips"},
			{Name: "inessential.com", URL: "https://inessential.com/xml/rss.xml", Description: "Brent Simmons’s weblog."},
			{Name: "tyler.io", URL: "https://tyler.io/feed/", Description: "tyler.io"},
		},
	},
}

// FindRelevant searches the curated feed database for feeds matching
// the given topic name and description. Returns up to 20 matching feeds.
func FindRelevant(topicName, description string) []Feed {
	query := strings.ToLower(topicName + " " + description)
	words := strings.Fields(query)

	// Filter to meaningful words (3+ chars)
	var keywords []string
	for _, w := range words {
		if len(w) >= 3 {
			keywords = append(keywords, w)
		}
	}

	if len(keywords) == 0 {
		return nil
	}

	seen := make(map[string]bool)
	var results []Feed

	for _, cat := range Categories {
		catLower := strings.ToLower(cat.Name)

		// Check if any keyword matches the category name
		catMatch := false
		for _, kw := range keywords {
			if strings.Contains(catLower, kw) {
				catMatch = true
				break
			}
		}

		if catMatch {
			// Include all feeds from matching category
			for _, f := range cat.Feeds {
				if !seen[f.URL] {
					seen[f.URL] = true
					results = append(results, f)
				}
			}
			continue
		}

		// Check individual feeds for keyword matches
		for _, f := range cat.Feeds {
			if seen[f.URL] {
				continue
			}
			feedText := strings.ToLower(f.Name + " " + f.Description)
			for _, kw := range keywords {
				if strings.Contains(feedText, kw) {
					seen[f.URL] = true
					results = append(results, f)
					break
				}
			}
		}
	}

	// Cap at 20 to keep prompts reasonable
	if len(results) > 20 {
		results = results[:20]
	}

	return results
}
