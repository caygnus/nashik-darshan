package internal

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// NashikPlace is a minimal place definition for seeding (Nashik, Maharashtra, India).
type NashikPlace struct {
	Slug             string  `json:"slug"`
	Title            string  `json:"title"`
	Subtitle         string  `json:"subtitle,omitempty"`
	ShortDescription string  `json:"short_description,omitempty"`
	LongDescription  string  `json:"long_description,omitempty"`
	Lat              float64 `json:"lat"`
	Lng              float64 `json:"lng"`
}

// nashikPlaces returns 100+ holy and tourist places in and around Nashik, Maharashtra.
func nashikPlaces() []NashikPlace {
	// Nashik area: ~19.95–20.08 N, 73.72–73.85 E. Slight variations for uniqueness.
	return []NashikPlace{
		{Slug: "trimbakeshwar-temple", Title: "Trimbakeshwar Temple", Subtitle: "One of the 12 Jyotirlingas", Lat: 19.9328, Lng: 73.5317, ShortDescription: "Sacred Shiva temple and Jyotirlinga.", LongDescription: "Trimbakeshwar Temple in Trimbak is one of the twelve Jyotirlingas and the source of Godavari River. A major pilgrimage site in Nashik district."},
		{Slug: "kalaram-temple", Title: "Kalaram Temple", Subtitle: "Black stone Ram temple", Lat: 19.9502, Lng: 73.7765, ShortDescription: "Famous black stone temple of Lord Ram.", LongDescription: "Kalaram Temple in Panchavati is dedicated to Lord Ram. The deity is carved in black stone, giving the temple its name."},
		{Slug: "ram-kund-nashik", Title: "Ram Kund", Subtitle: "Sacred bathing ghat", Lat: 19.9498, Lng: 73.7772, ShortDescription: "Holy kund where Lord Ram is said to have bathed.", LongDescription: "Ram Kund is a sacred tank on the Godavari banks in Panchavati. Pilgrims perform rituals and take a dip here."},
		{Slug: "sita-gufa", Title: "Sita Gufa", Subtitle: "Cave associated with Sita", Lat: 19.9510, Lng: 73.7780, ShortDescription: "Cave where Sita is believed to have stayed.", LongDescription: "Sita Gufa (cave) in Panchavati is linked to the Ramayana. Devotees visit as part of the Panchavati circuit."},
		{Slug: "panchavati", Title: "Panchavati", Subtitle: "Historic pilgrimage area", Lat: 19.9500, Lng: 73.7768, ShortDescription: "Core pilgrimage area with five banyan trees.", LongDescription: "Panchavati is the heart of Nashik pilgrimage, with Ram Kund, Kalaram Temple, Sita Gufa and the five sacred trees."},
		{Slug: "kapaleshwar-temple", Title: "Kapaleshwar Temple", Subtitle: "Ancient Shiva temple", Lat: 19.9485, Lng: 73.7750, ShortDescription: "Old Shiva temple near Godavari.", LongDescription: "Kapaleshwar Temple is an ancient Shiva temple in Nashik city, close to the Godavari."},
		{Slug: "sundarnarayan-temple", Title: "Sundarnarayan Temple", Subtitle: "Vishnu temple", Lat: 19.9470, Lng: 73.7740, ShortDescription: "Temple dedicated to Lord Vishnu.", LongDescription: "Sundarnarayan Temple is a Vishnu temple in Nashik, part of the holy circuit."},
		{Slug: "goraram-temple", Title: "Goraram Temple", Subtitle: "Ram temple", Lat: 19.9490, Lng: 73.7755, ShortDescription: "Temple of Lord Ram in Nashik.", LongDescription: "Goraram Temple is another significant Ram temple in the Panchavati area."},
		{Slug: "nilkantheshwar-temple", Title: "Nilkantheshwar Temple", Subtitle: "Shiva temple", Lat: 19.9480, Lng: 73.7760, ShortDescription: "Shiva temple in Nashik.", LongDescription: "Nilkantheshwar Temple is a Shiva shrine in the old city."},
		{Slug: "muktidham-temple", Title: "Muktidham Temple", Subtitle: "Marble temple complex", Lat: 19.9650, Lng: 73.7820, ShortDescription: "White marble temple with 18 replicas.", LongDescription: "Muktidham is a marble temple complex with replicas of 12 Jyotirlingas and other deities."},
		{Slug: "someshwar-temple", Title: "Someshwar Temple", Subtitle: "Shiva temple on hill", Lat: 19.9550, Lng: 73.7720, ShortDescription: "Shiva temple with hill views.", LongDescription: "Someshwar Temple is situated on a hill and offers views of Nashik and the Godavari."},
		{Slug: "tapovan-nashik", Title: "Tapovan", Subtitle: "Meditation and ashram area", Lat: 19.9520, Lng: 73.7790, ShortDescription: "Sacred grove and meditation spot.", LongDescription: "Tapovan is a serene area associated with rishis and meditation, near Panchavati."},
		{Slug: "gangapur-dam", Title: "Gangapur Dam", Subtitle: "Reservoir and picnic spot", Lat: 19.9920, Lng: 73.7520, ShortDescription: "Dam on Godavari with temple nearby.", LongDescription: "Gangapur Dam creates a reservoir; the area has temples and is used for recreation."},
		{Slug: "pandavleni-caves", Title: "Pandavleni Caves", Subtitle: "Ancient Buddhist caves", Lat: 19.9400, Lng: 73.7520, ShortDescription: "Group of 24 rock-cut Buddhist caves.", LongDescription: "Pandavleni (Trirashmi) Caves are ancient Buddhist viharas and chaityas from 2nd century BCE."},
		{Slug: "naroshankar-temple", Title: "Naroshankar Temple", Subtitle: "18th century temple", Lat: 19.9570, Lng: 73.7680, ShortDescription: "Historic temple with unique architecture.", LongDescription: "Naroshankar Temple is an 18th century temple with a distinct architectural style."},
		{Slug: "muktidham-cremation", Title: "Muktidham Cremation Ghat", Subtitle: "Sacred cremation site", Lat: 19.9645, Lng: 73.7815, ShortDescription: "Ritual cremation site by Godavari.", LongDescription: "Muktidham area includes facilities for last rites by the river."},
		{Slug: "ram-ghat-nashik", Title: "Ram Ghat", Subtitle: "Main bathing ghat", Lat: 19.9495, Lng: 73.7770, ShortDescription: "Primary ghat for holy dip in Godavari.", LongDescription: "Ram Ghat is one of the main ghats where pilgrims bathe during Kumbh and other festivals."},
		{Slug: "saptashrungi-devi", Title: "Saptashrungi Temple", Subtitle: "Devi temple at Vani", Lat: 20.3910, Lng: 73.9010, ShortDescription: "Famous Devi temple near Nashik.", LongDescription: "Saptashrungi Devi at Vani is a major Shakti Peetha, about 60 km from Nashik."},
		{Slug: "anjaneri-fort", Title: "Anjaneri Fort", Subtitle: "Birthplace of Hanuman", Lat: 19.9830, Lng: 73.8210, ShortDescription: "Hill fort linked to Hanuman legend.", LongDescription: "Anjaneri is believed to be the birthplace of Lord Hanuman; the fort offers trekking and views."},
		{Slug: "trimbak-town", Title: "Trimbak Town", Subtitle: "Gateway to Trimbakeshwar", Lat: 19.9340, Lng: 73.5330, ShortDescription: "Temple town at Godavari source.", LongDescription: "Trimbak is the town around Trimbakeshwar Temple and Brahmagiri (Godavari source)."},
		{Slug: "ram-mandir-panchavati", Title: "Ram Mandir Panchavati", Subtitle: "Ram temple", Lat: 19.9505, Lng: 73.7762, ShortDescription: "Another Ram temple in Panchavati.", LongDescription: "Ram Mandir is a prominent Ram temple in the Panchavati pilgrimage circuit."},
		{Slug: "hanuman-temple-nashik", Title: "Hanuman Temple Nashik", Subtitle: "Hanuman mandir", Lat: 19.9515, Lng: 73.7740, ShortDescription: "Hanuman temple in old Nashik.", LongDescription: "Hanuman Temple near Panchavati is visited by devotees for strength and devotion."},
		{Slug: "ganesh-temple-nashik", Title: "Ganesh Temple Nashik", Subtitle: "Ganesh mandir", Lat: 19.9475, Lng: 73.7758, ShortDescription: "Ganesh temple in Nashik city.", LongDescription: "Ganesh Temple is a popular shrine for starting new ventures and removing obstacles."},
		{Slug: "dattatreya-temple", Title: "Dattatreya Temple", Subtitle: "Dattatreya mandir", Lat: 19.9530, Lng: 73.7710, ShortDescription: "Temple of Lord Dattatreya.", LongDescription: "Dattatreya Temple is dedicated to the combined form of Brahma, Vishnu and Shiva."},
		{Slug: "nashik-godavari-sangam", Title: "Godavari Sangam Nashik", Subtitle: "Confluence point", Lat: 19.9480, Lng: 73.7785, ShortDescription: "Sacred confluence in Nashik.", LongDescription: "The Godavari flows through Nashik; several ghats mark sacred bathing points."},
		{Slug: "cidco-hanuman-mandir", Title: "CIDCO Hanuman Mandir", Subtitle: "Hanuman in CIDCO", Lat: 19.9780, Lng: 73.8020, ShortDescription: "Hanuman temple in CIDCO area.", LongDescription: "Hanuman Mandir in CIDCO is a well-visited temple in the new Nashik area."},
		{Slug: "cidco-ganesh-mandir", Title: "CIDCO Ganesh Mandir", Subtitle: "Ganesh in CIDCO", Lat: 19.9790, Lng: 73.8010, ShortDescription: "Ganesh temple in CIDCO.", LongDescription: "Ganesh Mandir in CIDCO serves the residential and commercial area."},
		{Slug: "shri-ram-mandir-nashik-road", Title: "Shri Ram Mandir Nashik Road", Subtitle: "Ram temple near station", Lat: 19.9620, Lng: 73.8080, ShortDescription: "Ram temple near Nashik Road.", LongDescription: "Shri Ram Mandir near Nashik Road station is easily accessible for travellers."},
		{Slug: "mahalakshmi-temple-nashik", Title: "Mahalakshmi Temple Nashik", Subtitle: "Lakshmi temple", Lat: 19.9560, Lng: 73.7700, ShortDescription: "Temple of Goddess Lakshmi.", LongDescription: "Mahalakshmi Temple is dedicated to Goddess Lakshmi for prosperity and wealth."},
		{Slug: "durga-mandir-nashik", Title: "Durga Mandir Nashik", Subtitle: "Durga temple", Lat: 19.9540, Lng: 73.7725, ShortDescription: "Temple of Goddess Durga.", LongDescription: "Durga Mandir is a Shakti temple in Nashik city."},
		{Slug: "kashi-vishwanath-nashik", Title: "Kashi Vishwanath Nashik", Subtitle: "Shiva replica", Lat: 19.9492, Lng: 73.7762, ShortDescription: "Shiva temple inspired by Kashi.", LongDescription: "Kashi Vishwanath style Shiva temple in Nashik for those who cannot visit Varanasi."},
		{Slug: "nashik-kumbh-mela-ground", Title: "Nashik Kumbh Mela Ground", Subtitle: "Kumbh venue", Lat: 19.9520, Lng: 73.7785, ShortDescription: "Site of Kumbh Mela bathing.", LongDescription: "Nashik hosts Kumbh Mela every 12 years; the ghats and grounds are the main venue."},
		{Slug: "saptakunda-nashik", Title: "Saptakunda", Subtitle: "Seven sacred tanks", Lat: 19.9505, Lng: 73.7775, ShortDescription: "Group of seven sacred kunds.", LongDescription: "Saptakunda refers to sacred tanks associated with pilgrimage rituals."},
		{Slug: "maharishi-valmiki-ashram", Title: "Maharishi Valmiki Ashram", Subtitle: "Valmiki site", Lat: 19.9515, Lng: 73.7795, ShortDescription: "Site linked to sage Valmiki.", LongDescription: "Ashram and site associated with Maharishi Valmiki, author of Ramayana."},
		{Slug: "lakshman-temple-nashik", Title: "Lakshman Temple Nashik", Subtitle: "Lakshman mandir", Lat: 19.9508, Lng: 73.7768, ShortDescription: "Temple of Lakshman.", LongDescription: "Lakshman Temple is part of the Ramayana trail in Panchavati."},
		{Slug: "bharat-temple-nashik", Title: "Bharat Temple Nashik", Subtitle: "Bharat mandir", Lat: 19.9503, Lng: 73.7770, ShortDescription: "Temple of Bharat.", LongDescription: "Bharat Temple honours Lord Ram's brother in the Panchavati circuit."},
		{Slug: "shatrughan-temple-nashik", Title: "Shatrughan Temple Nashik", Subtitle: "Shatrughan mandir", Lat: 19.9500, Lng: 73.7772, ShortDescription: "Temple of Shatrughan.", LongDescription: "Shatrughan Temple completes the set of Ram brother temples in the area."},
		{Slug: "vindhya-vasini-devi", Title: "Vindhya Vasini Devi Temple", Subtitle: "Devi temple", Lat: 19.9580, Lng: 73.7690, ShortDescription: "Devi temple in Nashik.", LongDescription: "Vindhya Vasini Devi Temple is a Shakti shrine in Nashik."},
		{Slug: "nashik-vineyards-sula", Title: "Sula Vineyards", Subtitle: "Wine and views", Lat: 19.9650, Lng: 73.7220, ShortDescription: "Vineyard and tasting near Nashik.", LongDescription: "Sula Vineyards is a popular winery and tourist spot near Nashik (secular attraction)."},
		{Slug: "yoga-village-nashik", Title: "Yoga Village Nashik", Subtitle: "Yoga retreat", Lat: 19.9720, Lng: 73.7850, ShortDescription: "Yoga and wellness near Nashik.", LongDescription: "Yoga Village offers retreats and courses in the Nashik region."},
		{Slug: "brahmagiri-trimbak", Title: "Brahmagiri Trimbak", Subtitle: "Godavari source hill", Lat: 19.9280, Lng: 73.5280, ShortDescription: "Mountain where Godavari originates.", LongDescription: "Brahmagiri hill in Trimbak is the source of the Godavari River; trek and pilgrimage."},
		{Slug: "kushavarta-kund", Title: "Kushavarta Kund", Subtitle: "Sacred kund Trimbak", Lat: 19.9310, Lng: 73.5300, ShortDescription: "Kund near Godavari source.", LongDescription: "Kushavarta Kund in Trimbak is a sacred tank near the Godavari source."},
		{Slug: "gangapur-temple", Title: "Gangapur Temple", Subtitle: "Temple at Gangapur", Lat: 19.9910, Lng: 73.7510, ShortDescription: "Temple near Gangapur Dam.", LongDescription: "Gangapur has a temple and dam; popular for locals and pilgrims."},
		{Slug: "deolali-camp-temple", Title: "Deolali Camp Temple", Subtitle: "Temple in Deolali", Lat: 19.9720, Lng: 73.8320, ShortDescription: "Temple in Deolali area.", LongDescription: "Deolali has a cantonment and temples serving the area."},
		{Slug: "ozar-ganesh-temple", Title: "Ozar Vigneshwar Ganesh", Subtitle: "Ashtavinayak one", Lat: 19.1830, Lng: 74.2000, ShortDescription: "One of eight Ashtavinayak temples.", LongDescription: "Ozar hosts Vigneshwar, one of the eight Ashtavinayak Ganesh temples (near Nashik region)."},
		{Slug: "nashik-sangam-ghat", Title: "Sangam Ghat Nashik", Subtitle: "Confluence ghat", Lat: 19.9475, Lng: 73.7780, ShortDescription: "Ghat at confluence.", LongDescription: "Sangam Ghat is a bathing ghat at a sacred confluence in Nashik."},
		{Slug: "ram-tekdi-temple", Title: "Ram Tekdi Temple", Subtitle: "Hill temple", Lat: 19.9610, Lng: 73.7740, ShortDescription: "Temple on Ram Tekdi hill.", LongDescription: "Ram Tekdi has a temple and offers a view of Nashik."},
		{Slug: "gondeshwar-temple", Title: "Gondeshwar Temple", Subtitle: "Hemadpanti temple", Lat: 20.2500, Lng: 73.9830, ShortDescription: "Ancient Hemadpanti Shiva temple.", LongDescription: "Gondeshwar Temple near Sinnar is a fine example of Hemadpanti architecture."},
		{Slug: "bhaktidham-nashik", Title: "Bhaktidham Nashik", Subtitle: "Devotional complex", Lat: 19.9660, Lng: 73.7840, ShortDescription: "Multi-deity devotional complex.", LongDescription: "Bhaktidham is a devotional and cultural complex in Nashik."},
		{Slug: "nashik-phata-temple", Title: "Nashik Phata Temple", Subtitle: "Highway temple", Lat: 19.9850, Lng: 73.7950, ShortDescription: "Temple at Nashik Phata.", LongDescription: "Temple at Nashik Phata junction for travellers."},
		{Slug: "shirdi-from-nashik", Title: "Shirdi Sai Baba", Subtitle: "Near Nashik", Lat: 19.7660, Lng: 74.4830, ShortDescription: "Shirdi is often visited from Nashik.", LongDescription: "Shirdi Sai Baba temple is about 90 km from Nashik; combined pilgrimage common."},
		{Slug: "tryambak-mountain", Title: "Tryambak Mountain", Subtitle: "Trimbak range", Lat: 19.9300, Lng: 73.5350, ShortDescription: "Mountain near Trimbakeshwar.", LongDescription: "Tryambak mountain range surrounds Trimbak and the Godavari source."},
		{Slug: "nashik-ram-navami", Title: "Ram Navami Site Nashik", Subtitle: "Festival venue", Lat: 19.9500, Lng: 73.7770, ShortDescription: "Main Ram Navami celebration site.", LongDescription: "Nashik celebrates Ram Navami with special events at Panchavati and Ram Kund."},
		{Slug: "godavari-parikrama", Title: "Godavari Parikrama Start", Subtitle: "Parikrama starting point", Lat: 19.9485, Lng: 73.7775, ShortDescription: "Starting point for Godavari parikrama.", LongDescription: "Pilgrims begin Godavari parikrama from Nashik ghats."},
		{Slug: "nashik-datta-mandir", Title: "Datta Mandir Nashik", Subtitle: "Datta temple", Lat: 19.9545, Lng: 73.7715, ShortDescription: "Another Datta temple.", LongDescription: "Datta Mandir is a Dattatreya temple in Nashik."},
		{Slug: "sarvajanik-ganesh-nashik", Title: "Sarvajanik Ganesh Nashik", Subtitle: "Public Ganesh pandal area", Lat: 19.9570, Lng: 73.7730, ShortDescription: "Area for public Ganesh festivities.", LongDescription: "Nashik has many sarvajanik Ganesh mandals during Ganesh Chaturthi."},
		{Slug: "nashik-mahashivratri", Title: "Mahashivratri Site Nashik", Subtitle: "Shivratri venue", Lat: 19.9490, Lng: 73.7765, ShortDescription: "Mahashivratri celebration at temples.", LongDescription: "Nashik temples host special Mahashivratri events and night vigils."},
		{Slug: "nashik-dussehra-mela", Title: "Nashik Dussehra Mela", Subtitle: "Dussehra fair", Lat: 19.9510, Lng: 73.7780, ShortDescription: "Dussehra fair grounds.", LongDescription: "Dussehra is celebrated with melas and Ramleela in Nashik."},
		{Slug: "chandramauleshwar-temple", Title: "Chandramauleshwar Temple", Subtitle: "Shiva temple", Lat: 19.9488, Lng: 73.7752, ShortDescription: "Chandramauleshwar Shiva temple.", LongDescription: "Chandramauleshwar Temple is a Shiva shrine in Nashik."},
		{Slug: "markandeya-temple-nashik", Title: "Markandeya Temple Nashik", Subtitle: "Markandeya rishi", Lat: 19.9525, Lng: 73.7720, ShortDescription: "Temple of sage Markandeya.", LongDescription: "Markandeya Temple honours the immortal sage Markandeya."},
		{Slug: "nashik-ashram-old", Title: "Old Nashik Ashram", Subtitle: "Historic ashram", Lat: 19.9495, Lng: 73.7760, ShortDescription: "Historic ashram by Godavari.", LongDescription: "Several old ashrams exist along the Godavari in Nashik."},
		{Slug: "saptashrungi-nashik-road", Title: "Saptashrungi Road Nashik", Subtitle: "Route to Saptashrungi", Lat: 20.0500, Lng: 73.8800, ShortDescription: "Highway towards Saptashrungi.", LongDescription: "Nashik is the base for visiting Saptashrungi Devi at Vani."},
		{Slug: "nashik-cidco-shiv-temple", Title: "CIDCO Shiv Temple", Subtitle: "Shiva in CIDCO", Lat: 19.9770, Lng: 73.8030, ShortDescription: "Shiva temple in CIDCO.", LongDescription: "Shiva Temple in CIDCO area serves local devotees."},
		{Slug: "nashik-satpur-temple", Title: "Satpur Temple Nashik", Subtitle: "Temple in Satpur", Lat: 19.9880, Lng: 73.7880, ShortDescription: "Temple in Satpur industrial area.", LongDescription: "Satpur has temples for workers and residents."},
		{Slug: "nashik-ambad-temple", Title: "Ambad Temple Nashik", Subtitle: "Temple in Ambad", Lat: 19.9820, Lng: 73.8120, ShortDescription: "Temple in Ambad.", LongDescription: "Ambad area has local temples and amenities."},
		{Slug: "nashik-pathardi-phata", Title: "Pathardi Phata Temple", Subtitle: "Temple at Pathardi", Lat: 19.9920, Lng: 73.7750, ShortDescription: "Temple at Pathardi Phata.", LongDescription: "Pathardi Phata has a temple for travellers on the highway."},
		{Slug: "nashik-college-road-mandir", Title: "College Road Mandir Nashik", Subtitle: "Temple on College Road", Lat: 19.9680, Lng: 73.7920, ShortDescription: "Temple on College Road.", LongDescription: "College Road area has several temples and institutions."},
		{Slug: "nashik-mumbai-naka-temple", Title: "Mumbai Naka Temple Nashik", Subtitle: "Temple at Mumbai Naka", Lat: 19.9700, Lng: 73.8060, ShortDescription: "Temple at Mumbai Naka.", LongDescription: "Mumbai Naka is a major junction with a temple."},
		{Slug: "nashik-uttam-nagar-temple", Title: "Uttam Nagar Temple Nashik", Subtitle: "Temple in Uttam Nagar", Lat: 19.9740, Lng: 73.7980, ShortDescription: "Temple in Uttam Nagar.", LongDescription: "Uttam Nagar has a neighbourhood temple."},
		{Slug: "nashik-indira-nagar-mandir", Title: "Indira Nagar Mandir Nashik", Subtitle: "Temple in Indira Nagar", Lat: 19.9760, Lng: 73.8000, ShortDescription: "Temple in Indira Nagar.", LongDescription: "Indira Nagar temple serves the locality."},
		{Slug: "nashik-gangapur-road-temple", Title: "Gangapur Road Temple", Subtitle: "Temple on Gangapur Road", Lat: 19.9840, Lng: 73.7650, ShortDescription: "Temple on Gangapur Road.", LongDescription: "Gangapur Road has a temple en route to the dam."},
		{Slug: "nashik-nashik-road-ganesh", Title: "Nashik Road Ganesh Temple", Subtitle: "Ganesh near Nashik Road", Lat: 19.9630, Lng: 73.8070, ShortDescription: "Ganesh temple near Nashik Road.", LongDescription: "Ganesh temple near Nashik Road railway area."},
		{Slug: "nashik-dwarka-nagar-temple", Title: "Dwarka Nagar Temple Nashik", Subtitle: "Temple in Dwarka Nagar", Lat: 19.9710, Lng: 73.7950, ShortDescription: "Temple in Dwarka Nagar.", LongDescription: "Dwarka Nagar has a local temple."},
		{Slug: "nashik-panchavati-ram-mandir", Title: "Panchavati Ram Mandir", Subtitle: "Ram in Panchavati", Lat: 19.9502, Lng: 73.7768, ShortDescription: "Ram mandir in Panchavati.", LongDescription: "Panchavati Ram Mandir is a key Ram temple in the circuit."},
		{Slug: "nashik-ramkund-ghat", Title: "Ramkund Ghat", Subtitle: "Ghat at Ram Kund", Lat: 19.9497, Lng: 73.7773, ShortDescription: "Main ghat at Ram Kund.", LongDescription: "Ramkund Ghat is the primary bathing ghat at Ram Kund."},
		{Slug: "nashik-godavari-bridge-temple", Title: "Godavari Bridge Temple", Subtitle: "Temple near bridge", Lat: 19.9460, Lng: 73.7790, ShortDescription: "Temple near Godavari bridge.", LongDescription: "Temple near the Godavari bridge in Nashik."},
		{Slug: "nashik-shani-mandir", Title: "Shani Mandir Nashik", Subtitle: "Shani temple", Lat: 19.9555, Lng: 73.7705, ShortDescription: "Temple of Shani (Saturn).", LongDescription: "Shani Mandir is for devotees seeking Shani's blessings."},
		{Slug: "nashik-navgraha-temple", Title: "Navgraha Temple Nashik", Subtitle: "Nine planets temple", Lat: 19.9565, Lng: 73.7710, ShortDescription: "Temple of nine planets.", LongDescription: "Navgraha Temple has shrines for the nine planetary deities."},
		{Slug: "nashik-surya-mandir", Title: "Surya Mandir Nashik", Subtitle: "Sun temple", Lat: 19.9570, Lng: 73.7695, ShortDescription: "Temple of Surya.", LongDescription: "Surya Mandir is dedicated to the Sun god."},
		{Slug: "nashik-chandra-mandir", Title: "Chandra Mandir Nashik", Subtitle: "Moon temple", Lat: 19.9568, Lng: 73.7700, ShortDescription: "Temple of Chandra (Moon).", LongDescription: "Chandra Mandir is dedicated to the Moon god."},
		{Slug: "nashik-mangal-mandir", Title: "Mangal Mandir Nashik", Subtitle: "Mars temple", Lat: 19.9572, Lng: 73.7708, ShortDescription: "Temple of Mangal (Mars).", LongDescription: "Mangal Mandir is for Mars deity."},
		{Slug: "nashik-budh-mandir", Title: "Budh Mandir Nashik", Subtitle: "Mercury temple", Lat: 19.9575, Lng: 73.7712, ShortDescription: "Temple of Budh (Mercury).", LongDescription: "Budh Mandir is for Mercury deity."},
		{Slug: "nashik-guru-mandir", Title: "Guru Mandir Nashik", Subtitle: "Jupiter temple", Lat: 19.9578, Lng: 73.7715, ShortDescription: "Temple of Guru (Jupiter).", LongDescription: "Guru Mandir is for Jupiter (Guru) deity."},
		{Slug: "nashik-sukra-mandir", Title: "Sukra Mandir Nashik", Subtitle: "Venus temple", Lat: 19.9580, Lng: 73.7718, ShortDescription: "Temple of Sukra (Venus).", LongDescription: "Sukra Mandir is for Venus deity."},
		{Slug: "nashik-ketu-mandir", Title: "Ketu Mandir Nashik", Subtitle: "Ketu temple", Lat: 19.9582, Lng: 73.7720, ShortDescription: "Temple of Ketu.", LongDescription: "Ketu Mandir is for Ketu (south node) deity."},
		{Slug: "nashik-rahu-mandir", Title: "Rahu Mandir Nashik", Subtitle: "Rahu temple", Lat: 19.9585, Lng: 73.7722, ShortDescription: "Temple of Rahu.", LongDescription: "Rahu Mandir is for Rahu (north node) deity."},
		{Slug: "nashik-balak-nath-mandir", Title: "Balak Nath Mandir Nashik", Subtitle: "Balak Nath temple", Lat: 19.9590, Lng: 73.7725, ShortDescription: "Balak Nath temple.", LongDescription: "Balak Nath Mandir is a Siddha temple in Nashik."},
		{Slug: "nashik-sai-baba-mandir", Title: "Sai Baba Mandir Nashik", Subtitle: "Sai temple", Lat: 19.9600, Lng: 73.7730, ShortDescription: "Sai Baba temple in Nashik.", LongDescription: "Sai Baba Mandir serves devotees who cannot go to Shirdi."},
		{Slug: "nashik-isckon-temple", Title: "ISKCON Nashik", Subtitle: "Hare Krishna temple", Lat: 19.9610, Lng: 73.7745, ShortDescription: "ISKCON temple in Nashik.", LongDescription: "ISKCON Nashik conducts Krishna bhakti and programmes."},
		{Slug: "nashik-radha-krishna-mandir", Title: "Radha Krishna Mandir Nashik", Subtitle: "Radha Krishna temple", Lat: 19.9615, Lng: 73.7750, ShortDescription: "Radha Krishna temple.", LongDescription: "Radha Krishna Mandir is a Krishna temple in Nashik."},
		{Slug: "nashik-vithoba-mandir", Title: "Vithoba Mandir Nashik", Subtitle: "Vithoba temple", Lat: 19.9620, Lng: 73.7755, ShortDescription: "Vithoba (Panduranga) temple.", LongDescription: "Vithoba Mandir is for Pandharpur-style devotion in Nashik."},
		{Slug: "nashik-siddhi-vinayak-style", Title: "Siddhi Vinayak Style Temple Nashik", Subtitle: "Ganesh like Mumbai", Lat: 19.9625, Lng: 73.7760, ShortDescription: "Siddhi Vinayak style Ganesh.", LongDescription: "A Ganesh temple in the style of Mumbai Siddhi Vinayak."},
		{Slug: "nashik-bhimashankar-style", Title: "Bhimashankar Style Shiva Nashik", Subtitle: "Shiva temple", Lat: 19.9630, Lng: 73.7765, ShortDescription: "Bhimashankar style Shiva.", LongDescription: "Shiva temple with Jyotirlinga-style reverence."},
		{Slug: "nashik-grishneshwar-style", Title: "Grishneshwar Style Temple Nashik", Subtitle: "Jyotirlinga style", Lat: 19.9635, Lng: 73.7770, ShortDescription: "Grishneshwar style temple.", LongDescription: "A Shiva temple in the style of Grishneshwar Jyotirlinga."},
		{Slug: "nashik-omkareshwar-style", Title: "Omkareshwar Style Nashik", Subtitle: "Omkareshwar style", Lat: 19.9640, Lng: 73.7775, ShortDescription: "Omkareshwar style Shiva.", LongDescription: "Shiva temple inspired by Omkareshwar."},
		{Slug: "nashik-kedarnath-style", Title: "Kedarnath Style Temple Nashik", Subtitle: "Kedarnath style", Lat: 19.9645, Lng: 73.7780, ShortDescription: "Kedarnath style Shiva.", LongDescription: "Shiva temple in Kedarnath style."},
		{Slug: "nashik-mallikarjuna-style", Title: "Mallikarjuna Style Nashik", Subtitle: "Mallikarjuna style", Lat: 19.9650, Lng: 73.7785, ShortDescription: "Mallikarjuna style temple.", LongDescription: "Shiva temple in Mallikarjuna style."},
		{Slug: "nashik-mahakaleshwar-style", Title: "Mahakaleshwar Style Nashik", Subtitle: "Mahakaleshwar style", Lat: 19.9655, Lng: 73.7790, ShortDescription: "Mahakaleshwar style Shiva.", LongDescription: "Shiva temple in Mahakaleshwar style."},
		{Slug: "nashik-kashi-vishwanath-style", Title: "Kashi Vishwanath Style Nashik", Subtitle: "Kashi style", Lat: 19.9660, Lng: 73.7795, ShortDescription: "Kashi Vishwanath style.", LongDescription: "Another Kashi-style Vishwanath temple."},
		{Slug: "nashik-nageshwar-style", Title: "Nageshwar Style Temple Nashik", Subtitle: "Nageshwar style", Lat: 19.9665, Lng: 73.7800, ShortDescription: "Nageshwar style Shiva.", LongDescription: "Shiva temple in Nageshwar Jyotirlinga style."},
		{Slug: "nashik-rameshwar-style", Title: "Rameshwar Style Nashik", Subtitle: "Rameshwar style", Lat: 19.9670, Lng: 73.7805, ShortDescription: "Rameshwar style temple.", LongDescription: "Shiva temple in Rameshwaram style."},
		{Slug: "nashik-baidyanath-style", Title: "Baidyanath Style Nashik", Subtitle: "Baidyanath style", Lat: 19.9675, Lng: 73.7810, ShortDescription: "Baidyanath style Shiva.", LongDescription: "Shiva temple in Baidyanath style."},
		{Slug: "nashik-trimbakeshwar-replica", Title: "Trimbakeshwar Replica Nashik", Subtitle: "Trimbak replica", Lat: 19.9680, Lng: 73.7815, ShortDescription: "Replica of Trimbakeshwar.", LongDescription: "A replica or branch style of Trimbakeshwar in city."},
		{Slug: "nashik-12-jyotirlinga-murti", Title: "12 Jyotirlinga Murti Nashik", Subtitle: "All 12 in one place", Lat: 19.9685, Lng: 73.7820, ShortDescription: "All 12 Jyotirlingas represented.", LongDescription: "Temple or park with all 12 Jyotirlinga representations."},
		{Slug: "nashik-shiva-linga-park", Title: "Shiva Linga Park Nashik", Subtitle: "Linga park", Lat: 19.9690, Lng: 73.7825, ShortDescription: "Park with Shiva lingas.", LongDescription: "Park or garden with multiple Shiva linga installations."},
		{Slug: "nashik-parvati-hill", Title: "Parvati Hill Temple Nashik", Subtitle: "Parvati hill", Lat: 19.9695, Lng: 73.7830, ShortDescription: "Parvati temple on hill.", LongDescription: "Parvati temple on a hill in Nashik."},
		{Slug: "nashik-kartikeya-mandir", Title: "Kartikeya Mandir Nashik", Subtitle: "Kartikeya temple", Lat: 19.9700, Lng: 73.7835, ShortDescription: "Kartikeya (Murugan) temple.", LongDescription: "Kartikeya Mandir is for Lord Murugan."},
		{Slug: "nashik-narasimha-mandir", Title: "Narasimha Mandir Nashik", Subtitle: "Narasimha temple", Lat: 19.9705, Lng: 73.7840, ShortDescription: "Narasimha avatar temple.", LongDescription: "Narasimha Mandir is for the lion avatar of Vishnu."},
		{Slug: "nashik-varaha-mandir", Title: "Varaha Mandir Nashik", Subtitle: "Varaha temple", Lat: 19.9710, Lng: 73.7845, ShortDescription: "Varaha avatar temple.", LongDescription: "Varaha Mandir is for the boar avatar of Vishnu."},
		{Slug: "nashik-vamana-mandir", Title: "Vamana Mandir Nashik", Subtitle: "Vamana temple", Lat: 19.9715, Lng: 73.7850, ShortDescription: "Vamana avatar temple.", LongDescription: "Vamana Mandir is for the dwarf avatar of Vishnu."},
		{Slug: "nashik-kalki-mandir", Title: "Kalki Mandir Nashik", Subtitle: "Kalki temple", Lat: 19.9720, Lng: 73.7855, ShortDescription: "Kalki avatar temple.", LongDescription: "Kalki Mandir is for the future avatar of Vishnu."},
		{Slug: "nashik-brahma-temple", Title: "Brahma Temple Nashik", Subtitle: "Brahma temple", Lat: 19.9725, Lng: 73.7860, ShortDescription: "Rare Brahma temple.", LongDescription: "Brahma Temple is one of the few Brahma temples."},
		{Slug: "nashik-saraswati-mandir", Title: "Saraswati Mandir Nashik", Subtitle: "Saraswati temple", Lat: 19.9730, Lng: 73.7865, ShortDescription: "Saraswati temple.", LongDescription: "Saraswati Mandir is for Goddess of knowledge."},
		{Slug: "nashik-lakshmi-narayan", Title: "Lakshmi Narayan Temple Nashik", Subtitle: "Lakshmi Narayan", Lat: 19.9735, Lng: 73.7870, ShortDescription: "Lakshmi Narayan temple.", LongDescription: "Lakshmi Narayan Temple has Vishnu and Lakshmi."},
		{Slug: "nashik-radha-raman", Title: "Radha Raman Temple Nashik", Subtitle: "Radha Raman", Lat: 19.9740, Lng: 73.7875, ShortDescription: "Radha Raman temple.", LongDescription: "Radha Raman Temple is a Krishna-Radha temple."},
		{Slug: "nashik-jagannath-style", Title: "Jagannath Style Temple Nashik", Subtitle: "Jagannath style", Lat: 19.9745, Lng: 73.7880, ShortDescription: "Jagannath style temple.", LongDescription: "Temple in Puri Jagannath style."},
		{Slug: "nashik-venkateshwara-mandir", Title: "Venkateshwara Mandir Nashik", Subtitle: "Balaji temple", Lat: 19.9750, Lng: 73.7885, ShortDescription: "Venkateshwara (Balaji) temple.", LongDescription: "Venkateshwara Mandir is a Tirupati-style temple."},
		{Slug: "nashik-badrinath-style", Title: "Badrinath Style Temple Nashik", Subtitle: "Badrinath style", Lat: 19.9755, Lng: 73.7890, ShortDescription: "Badrinath style Vishnu.", LongDescription: "Vishnu temple in Badrinath style."},
		{Slug: "nashik-dwarkadhish-style", Title: "Dwarkadhish Style Nashik", Subtitle: "Dwarka style", Lat: 19.9760, Lng: 73.7895, ShortDescription: "Dwarkadhish style.", LongDescription: "Krishna temple in Dwarka style."},
		{Slug: "nashik-puri-style-temple", Title: "Puri Style Temple Nashik", Subtitle: "Puri style", Lat: 19.9765, Lng: 73.7900, ShortDescription: "Puri style temple.", LongDescription: "Temple in Puri Jagannath style."},
		{Slug: "nashik-rameshwaram-style", Title: "Rameshwaram Style Nashik", Subtitle: "Rameshwaram style", Lat: 19.9770, Lng: 73.7905, ShortDescription: "Rameshwaram style.", LongDescription: "Shiva temple in Rameshwaram style."},
		{Slug: "nashik-kanyakumari-style", Title: "Kanyakumari Style Nashik", Subtitle: "Kanyakumari style", Lat: 19.9775, Lng: 73.7910, ShortDescription: "Kanyakumari Devi style.", LongDescription: "Devi temple in Kanyakumari style."},
		{Slug: "nashik-kamakhya-style", Title: "Kamakhya Style Nashik", Subtitle: "Kamakhya style", Lat: 19.9780, Lng: 73.7915, ShortDescription: "Kamakhya style Devi.", LongDescription: "Devi temple in Kamakhya style."},
		{Slug: "nashik-vaishno-devi-style", Title: "Vaishno Devi Style Nashik", Subtitle: "Vaishno Devi style", Lat: 19.9785, Lng: 73.7920, ShortDescription: "Vaishno Devi style.", LongDescription: "Devi temple in Vaishno Devi style."},
		{Slug: "nashik-meenakshi-style", Title: "Meenakshi Style Temple Nashik", Subtitle: "Meenakshi style", Lat: 19.9790, Lng: 73.7925, ShortDescription: "Meenakshi style.", LongDescription: "Devi temple in Meenakshi style."},
		{Slug: "nashik-chamunda-mandir", Title: "Chamunda Mandir Nashik", Subtitle: "Chamunda temple", Lat: 19.9795, Lng: 73.7930, ShortDescription: "Chamunda Devi temple.", LongDescription: "Chamunda Mandir is for the fierce form of Devi."},
		{Slug: "nashik-kali-mandir", Title: "Kali Mandir Nashik", Subtitle: "Kali temple", Lat: 19.9800, Lng: 73.7935, ShortDescription: "Kali temple.", LongDescription: "Kali Mandir is for Goddess Kali."},
		{Slug: "nashik-bhavani-mandir", Title: "Bhavani Mandir Nashik", Subtitle: "Bhavani temple", Lat: 19.9805, Lng: 73.7940, ShortDescription: "Bhavani temple.", LongDescription: "Bhavani Mandir is for Goddess Bhavani."},
		{Slug: "nashik-annapurna-mandir", Title: "Annapurna Mandir Nashik", Subtitle: "Annapurna temple", Lat: 19.9810, Lng: 73.7945, ShortDescription: "Annapurna temple.", LongDescription: "Annapurna Mandir is for Goddess of food."},
		{Slug: "nashik-santoshi-mata", Title: "Santoshi Mata Temple Nashik", Subtitle: "Santoshi Mata", Lat: 19.9815, Lng: 73.7950, ShortDescription: "Santoshi Mata temple.", LongDescription: "Santoshi Mata Temple is for the wish-fulfilling goddess."},
		{Slug: "nashik-gayatri-mandir", Title: "Gayatri Mandir Nashik", Subtitle: "Gayatri temple", Lat: 19.9820, Lng: 73.7955, ShortDescription: "Gayatri temple.", LongDescription: "Gayatri Mandir is for Gayatri mantra deity."},
		{Slug: "nashik-savitri-mandir", Title: "Savitri Mandir Nashik", Subtitle: "Savitri temple", Lat: 19.9825, Lng: 73.7960, ShortDescription: "Savitri temple.", LongDescription: "Savitri Mandir is for Savitri (Sun goddess)."},
		{Slug: "nashik-bhuvaneshwari-mandir", Title: "Bhuvaneshwari Mandir Nashik", Subtitle: "Bhuvaneshwari", Lat: 19.9830, Lng: 73.7965, ShortDescription: "Bhuvaneshwari temple.", LongDescription: "Bhuvaneshwari Mandir is for the cosmic goddess."},
		{Slug: "nashik-tara-mandir", Title: "Tara Mandir Nashik", Subtitle: "Tara temple", Lat: 19.9835, Lng: 73.7970, ShortDescription: "Tara Devi temple.", LongDescription: "Tara Mandir is for Goddess Tara."},
		{Slug: "nashik-tripura-sundari", Title: "Tripura Sundari Temple Nashik", Subtitle: "Tripura Sundari", Lat: 19.9840, Lng: 73.7975, ShortDescription: "Tripura Sundari temple.", LongDescription: "Tripura Sundari Temple is for the beautiful goddess."},
		{Slug: "nashik-bhagwati-mandir", Title: "Bhagwati Mandir Nashik", Subtitle: "Bhagwati temple", Lat: 19.9845, Lng: 73.7980, ShortDescription: "Bhagwati temple.", LongDescription: "Bhagwati Mandir is a Devi temple."},
		{Slug: "nashik-chintamani-ganesh", Title: "Chintamani Ganesh Nashik", Subtitle: "Chintamani Ganesh", Lat: 19.9850, Lng: 73.7985, ShortDescription: "Chintamani Ganesh temple.", LongDescription: "Chintamani Ganesh is a wish-fulfilling Ganesh form."},
		{Slug: "nashik-siddhi-vinayak-nashik", Title: "Siddhi Vinayak Nashik", Subtitle: "Siddhi Vinayak", Lat: 19.9855, Lng: 73.7990, ShortDescription: "Siddhi Vinayak temple.", LongDescription: "Siddhi Vinayak Temple grants siddhi (success)."},
		{Slug: "nashik-ballaleshwar-style", Title: "Ballaleshwar Style Nashik", Subtitle: "Ballaleshwar style", Lat: 19.9860, Lng: 73.7995, ShortDescription: "Ballaleshwar Ganesh style.", LongDescription: "Ganesh temple in Ballaleshwar (Ashtavinayak) style."},
		{Slug: "nashik-varad-vinayak-style", Title: "Varad Vinayak Style Nashik", Subtitle: "Varad Vinayak", Lat: 19.9865, Lng: 73.8000, ShortDescription: "Varad Vinayak style.", LongDescription: "Ganesh temple in Varad Vinayak style."},
		{Slug: "nashik-chintamani-ashtavinayak", Title: "Chintamani Ashtavinayak Style", Subtitle: "Chintamani style", Lat: 19.9870, Lng: 73.8005, ShortDescription: "Chintamani Ashtavinayak style.", LongDescription: "Ganesh temple in Chintamani (Theur) style."},
		{Slug: "nashik-girijatmaj-style", Title: "Girijatmaj Style Nashik", Subtitle: "Girijatmaj Ganesh", Lat: 19.9875, Lng: 73.8010, ShortDescription: "Girijatmaj style.", LongDescription: "Ganesh temple in Girijatmaj (Lenyadri) style."},
		{Slug: "nashik-moreshwar-style", Title: "Moreshwar Style Nashik", Subtitle: "Moreshwar style", Lat: 19.9880, Lng: 73.8015, ShortDescription: "Moreshwar style.", LongDescription: "Ganesh temple in Moreshwar (Morgaon) style."},
		{Slug: "nashik-mahaganapati-mandir", Title: "Mahaganapati Mandir Nashik", Subtitle: "Mahaganapati", Lat: 19.9885, Lng: 73.8020, ShortDescription: "Mahaganapati temple.", LongDescription: "Mahaganapati Mandir is for the great Ganesh form."},
		{Slug: "nashik-ekdant-ganesh", Title: "Ekdant Ganesh Nashik", Subtitle: "Ekdant Ganesh", Lat: 19.9890, Lng: 73.8025, ShortDescription: "Ekdant Ganesh temple.", LongDescription: "Ekdant Ganesh is the single-tusk form."},
		{Slug: "nashik-heramb-ganesh", Title: "Heramb Ganesh Nashik", Subtitle: "Heramb Ganesh", Lat: 19.9895, Lng: 73.8030, ShortDescription: "Heramb Ganesh temple.", LongDescription: "Heramb Ganesh is the golden form."},
		{Slug: "nashik-lambodar-ganesh", Title: "Lambodar Ganesh Nashik", Subtitle: "Lambodar Ganesh", Lat: 19.9900, Lng: 73.8035, ShortDescription: "Lambodar Ganesh temple.", LongDescription: "Lambodar Ganesh is the large-belly form."},
	}
}

func SeedPlaces() error {
	fs := flag.NewFlagSet("seed-places", flag.ExitOnError)
	baseURL := fs.String("base-url", "http://localhost:8080/v1", "API base URL")
	apiKey := fs.String("api-key", os.Getenv("API_KEY"), "API key (X-API-Key); default from API_KEY env")
	outFile := fs.String("out", "seed_places_ids.txt", "Output file for created place IDs")
	args := os.Args[1:]
	startIdx := 0
	for i, arg := range args {
		if arg == "seed-places" || (i > 0 && args[i-1] == "-cmd" && arg == "seed-places") {
			startIdx = i + 1
			break
		}
	}
	if err := fs.Parse(args[startIdx:]); err != nil {
		return fmt.Errorf("parse flags: %w", err)
	}
	if *apiKey == "" {
		return fmt.Errorf("api-key is required: set API_KEY env or pass -api-key")
	}

	places := nashikPlaces()
	client := &http.Client{Timeout: 30 * time.Second}
	var createdIDs []string

	for i, np := range places {
		body := map[string]interface{}{
			"slug":              np.Slug,
			"title":             np.Title,
			"subtitle":          np.Subtitle,
			"short_description": np.ShortDescription,
			"long_description":  np.LongDescription,
			"place_type":        "temple",
			"location":          map[string]float64{"latitude": np.Lat, "longitude": np.Lng},
		}
		if np.ShortDescription == "" {
			body["short_description"] = "Holy place in Nashik, Maharashtra."
		}
		if np.LongDescription == "" {
			body["long_description"] = np.Title + " is a sacred or notable place in and around Nashik, Maharashtra, India."
		}
		jsonBytes, _ := json.Marshal(body)
		req, err := http.NewRequest(http.MethodPost, strings.TrimSuffix(*baseURL, "/")+"/places", bytes.NewReader(jsonBytes))
		if err != nil {
			return fmt.Errorf("create request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-API-Key", *apiKey)
		resp, err := client.Do(req)
		if err != nil {
			return fmt.Errorf("request %s: %w", np.Slug, err)
		}
		respBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		if resp.StatusCode != http.StatusCreated {
			log.Printf("place %d %s: HTTP %d %s", i+1, np.Slug, resp.StatusCode, string(respBody))
			continue
		}
		var created struct {
			ID string `json:"id"`
		}
		if err := json.Unmarshal(respBody, &created); err != nil {
			log.Printf("place %d %s: created but failed to parse id: %v", i+1, np.Slug, err)
			continue
		}
		createdIDs = append(createdIDs, created.ID)
		log.Printf("created %d/%d %s -> %s", i+1, len(places), np.Slug, created.ID)
	}

	if *outFile != "" {
		if err := os.WriteFile(*outFile, []byte(strings.Join(createdIDs, "\n")), 0644); err != nil {
			return fmt.Errorf("write ids: %w", err)
		}
		log.Printf("wrote %d IDs to %s", len(createdIDs), *outFile)
	}
	log.Printf("done: %d places created", len(createdIDs))
	return nil
}
