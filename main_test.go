package main

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCanonicalURLs(t *testing.T) {
	require.Equal(t, sampleContentCanonicalized, canonicalizeURLs(sampleContent))
}

func TestMinimizeContent(t *testing.T) {
	require.Equal(t, sampleContentMinimized, minimizeContent(sampleContent))
}

func TestRenderLayout(t *testing.T) {
	// Just a very basic check that things work without erroring
	_, err := renderLayout("sequences.tmpl.html", "a title", "some content", time.Now())
	require.NoError(t, err)
}

func TestShouldRetryStatusCode(t *testing.T) {
	require.True(t, shouldRetryStatusCode(http.StatusTooManyRequests))
	require.True(t, shouldRetryStatusCode(http.StatusInternalServerError))

	// Conflict is returned by a Spring '83 implementation in cases where a
	// newer version of a board has already been posted, so if we encounter
	// this, consider it a success and stop retrying.
	require.False(t, shouldRetryStatusCode(http.StatusConflict))
}

//nolint:lll
const sampleContent = `<p>You&rsquo;ll have to give me a break on photo quality for this one &ndash; it&rsquo;s hard getting something good through the foggy glass of a plane window.</p>

<p>This is <a href="https://en.wikipedia.org/wiki/Mount_Rainier">Mount Rainier</a>, the tallest mountain in Washington state and the Cascade mountain range, and also one of the most dangerous volcanoes in the world. It&rsquo;s on the list of <a href="https://en.wikipedia.org/wiki/Decade_Volcanoes">Decade Volcanoes</a> thanks to its history of large, destructive eruptions and near proximity to a dense populzation zone. Wikipedia almost notes that it&rsquo;s the most topologically prominent peak in the contiguous US, dwarfing everything else around it and having quite a striking effect on the eye.</p>

<p>I just landed in Seattle. It&rsquo;s colder than expected. Like colder than it rightfully should be in any west coast city. Luckily, I learned from <a href="/nanoglyphs/033-heroku#new-york">my mistake in New York</a> and came equipped with a variety of cold weather gear this time around. I haven&rsquo;t had a chance to do much yet besides check into my hotel and head over to the flagship Amazon Go store, which was quite busy, but appeared to be about 5% shoppers, and 95% senior Amazon staff chatting in small circles, lauding each other on their own ingenuity. Still, it was nice seeing a downtown that&rsquo;d regained some of its lost vibrancy.</p>

<p>I got a coffee, along with a note saying that Amazon is &ldquo;working on my receipt&rdquo;, but nothing since. I suspect it might be a Mechanical Turk who ends up piecing together my bill from video rather than the finely tuned neural nets of a hyper-sophisticated ML cluster, but I might be a cynic. On my way out, someone handed me a free banana from a cart parked next to a geodesic dome.</p>

<p>Last weekend I wrote a <a href="https://github.com/brandur/spring83-keygen">Spring &lsquo;83 key generator</a>, and on the flight got maybe halfway to a working server implementation. Tomorrow, more Seattle, more Spring &lsquo;83, and work time spent on SSO and polish on a forthcoming metrics product for Bridge.</p>


<img src="/photographs/sequences/030_large.jpg"
    srcset="/photographs/sequences/030_large@2x.jpg 2x, /photographs/sequences/030_large.jpg 1x">

<img src="/photographs/sequences/030b_large.jpg"
    srcset="/photographs/sequences/030b_large@2x.jpg 2x, /photographs/sequences/030b_large.jpg 1x">

<img src="/photographs/sequences/030c_large.jpg"
    srcset="/photographs/sequences/030c_large@2x.jpg 2x, /photographs/sequences/030c_large.jpg 1x">`

//nolint:lll
const sampleContentCanonicalized = `<p>You&rsquo;ll have to give me a break on photo quality for this one &ndash; it&rsquo;s hard getting something good through the foggy glass of a plane window.</p>

<p>This is <a href="https://en.wikipedia.org/wiki/Mount_Rainier">Mount Rainier</a>, the tallest mountain in Washington state and the Cascade mountain range, and also one of the most dangerous volcanoes in the world. It&rsquo;s on the list of <a href="https://en.wikipedia.org/wiki/Decade_Volcanoes">Decade Volcanoes</a> thanks to its history of large, destructive eruptions and near proximity to a dense populzation zone. Wikipedia almost notes that it&rsquo;s the most topologically prominent peak in the contiguous US, dwarfing everything else around it and having quite a striking effect on the eye.</p>

<p>I just landed in Seattle. It&rsquo;s colder than expected. Like colder than it rightfully should be in any west coast city. Luckily, I learned from <a href="https://brandur.org/nanoglyphs/033-heroku#new-york">my mistake in New York</a> and came equipped with a variety of cold weather gear this time around. I haven&rsquo;t had a chance to do much yet besides check into my hotel and head over to the flagship Amazon Go store, which was quite busy, but appeared to be about 5% shoppers, and 95% senior Amazon staff chatting in small circles, lauding each other on their own ingenuity. Still, it was nice seeing a downtown that&rsquo;d regained some of its lost vibrancy.</p>

<p>I got a coffee, along with a note saying that Amazon is &ldquo;working on my receipt&rdquo;, but nothing since. I suspect it might be a Mechanical Turk who ends up piecing together my bill from video rather than the finely tuned neural nets of a hyper-sophisticated ML cluster, but I might be a cynic. On my way out, someone handed me a free banana from a cart parked next to a geodesic dome.</p>

<p>Last weekend I wrote a <a href="https://github.com/brandur/spring83-keygen">Spring &lsquo;83 key generator</a>, and on the flight got maybe halfway to a working server implementation. Tomorrow, more Seattle, more Spring &lsquo;83, and work time spent on SSO and polish on a forthcoming metrics product for Bridge.</p>


<img src="https://brandur.org/photographs/sequences/030_large.jpg"
    srcset="/photographs/sequences/030_large@2x.jpg 2x, /photographs/sequences/030_large.jpg 1x">

<img src="https://brandur.org/photographs/sequences/030b_large.jpg"
    srcset="/photographs/sequences/030b_large@2x.jpg 2x, /photographs/sequences/030b_large.jpg 1x">

<img src="https://brandur.org/photographs/sequences/030c_large.jpg"
    srcset="/photographs/sequences/030c_large@2x.jpg 2x, /photographs/sequences/030c_large.jpg 1x">`

//nolint:lll
const sampleContentMinimized = `<p>You???ll have to give me a break on photo quality for this one ??? it???s hard getting something good through the foggy glass of a plane window.</p><p>This is <a href="https://en.wikipedia.org/wiki/Mount_Rainier">Mount Rainier</a>, the tallest mountain in Washington state and the Cascade mountain range, and also one of the most dangerous volcanoes in the world. It???s on the list of <a href="https://en.wikipedia.org/wiki/Decade_Volcanoes">Decade Volcanoes</a> thanks to its history of large, destructive eruptions and near proximity to a dense populzation zone. Wikipedia almost notes that it???s the most topologically prominent peak in the contiguous US, dwarfing everything else around it and having quite a striking effect on the eye.</p><p>I just landed in Seattle. It???s colder than expected. Like colder than it rightfully should be in any west coast city. Luckily, I learned from <a href="/nanoglyphs/033-heroku#new-york">my mistake in New York</a> and came equipped with a variety of cold weather gear this time around. I haven???t had a chance to do much yet besides check into my hotel and head over to the flagship Amazon Go store, which was quite busy, but appeared to be about 5% shoppers, and 95% senior Amazon staff chatting in small circles, lauding each other on their own ingenuity. Still, it was nice seeing a downtown that???d regained some of its lost vibrancy.</p><p>I got a coffee, along with a note saying that Amazon is ???working on my receipt???, but nothing since. I suspect it might be a Mechanical Turk who ends up piecing together my bill from video rather than the finely tuned neural nets of a hyper-sophisticated ML cluster, but I might be a cynic. On my way out, someone handed me a free banana from a cart parked next to a geodesic dome.</p><p>Last weekend I wrote a <a href="https://github.com/brandur/spring83-keygen">Spring ???83 key generator</a>, and on the flight got maybe halfway to a working server implementation. Tomorrow, more Seattle, more Spring ???83, and work time spent on SSO and polish on a forthcoming metrics product for Bridge.</p><img src="/photographs/sequences/030_large.jpg"><img src="/photographs/sequences/030b_large.jpg"><img src="/photographs/sequences/030c_large.jpg">`
