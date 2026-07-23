// SPDX-License-Identifier: AGPL-3.0-only
// SPDX-FileCopyrightText: 2026 the Mosaic authors
// Linking exception: see LICENSE-EXCEPTION.

package playback

import (
	"reflect"
	"testing"
)

// theHindiFirstRelease is the fixture ADR 0050 was written around: a 4K HEVC
// Matroska with four E-AC3 tracks whose first is Hindi. It is the case that
// proves storing only the Part's scalar columns would have been a regression —
// nothing in Container/VideoCodec/AudioCodec can say which of these four plays.
var theHindiFirstRelease = MediaInfo{
	Container: "matroska",
	SizeBytes: 34_336_638_566,
	Video: VideoTrack{
		Index: 0, Codec: "hevc", Width: 3840, Height: 2160,
		Profile: "Main 10", PixelFmt: "yuv420p10le", HDRFormat: "hdr10",
	},
	Audio: []AudioTrack{
		{Index: 1, Codec: "eac3", Channels: 6, Language: "hin", Default: true},
		{Index: 2, Codec: "eac3", Channels: 6, Language: "tam"},
		{Index: 3, Codec: "eac3", Channels: 6, Language: "tel"},
		{Index: 4, Codec: "eac3", Channels: 6, Language: "eng"},
	},
	Subtitles: []SubtitleTrack{
		{Index: 5, Codec: "subrip", Language: "eng"},
		{Index: 6, Codec: "subrip", Language: "hin", Forced: true},
	},
}

// TestProbeSurvivesStorage is the property the whole slice rests on: a plan made
// from a stored probe must be identical to one made from a fresh probe. If the
// round trip loses anything, the second play of a release quietly differs from
// the first — and it would differ in the audio track, which is the part a viewer
// notices immediately and a test would not.
func TestProbeSurvivesStorage(t *testing.T) {
	raw, err := Encode(theHindiFirstRelease)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	got, ok := Decode(raw)
	if !ok {
		t.Fatal("Decode rejected a document it had just written")
	}
	if !reflect.DeepEqual(got, theHindiFirstRelease) {
		t.Errorf("round trip changed the probe:\n got %+v\nwant %+v", got, theHindiFirstRelease)
	}

	fresh := Decide(theHindiFirstRelease, DefaultBrowserCodecs, nil)
	stored := Decide(got, DefaultBrowserCodecs, nil)
	if !reflect.DeepEqual(fresh, stored) {
		t.Errorf("a stored probe produced a different plan:\n stored %+v\n fresh  %+v", stored, fresh)
	}
	// And specifically: still English, not the default-flagged Hindi track.
	if stored.AudioLanguage != "eng" {
		t.Errorf("audio language = %q, want eng — the stored probe lost track selection", stored.AudioLanguage)
	}
}

// TestDecodeRejectsWhatItCannotTrust covers every input that must produce a
// re-probe rather than a guess. Re-probing costs one ffprobe run; guessing costs
// a wrong plan, and the wrong plan is silent.
func TestDecodeRejectsWhatItCannotTrust(t *testing.T) {
	cases := map[string][]byte{
		"absent":         nil,
		"empty":          {},
		"not json":       []byte(`not json at all`),
		"not an object":  []byte(`[1,2,3]`),
		"no version":     []byte(`{"container":"matroska"}`),
		"future version": []byte(`{"v":99,"container":"matroska"}`),
	}
	for name, raw := range cases {
		if _, ok := Decode(raw); ok {
			t.Errorf("Decode accepted %s", name)
		}
	}
}

// TestSummaryAudioCodecNamesThePlayedTrack pins the column that feeds candidate
// ranking (ADR 0048) to the track the plan would actually choose. The first
// track and the default-flagged track are both the wrong answer here, and both
// are what a naive summary would pick.
func TestSummaryAudioCodecNamesThePlayedTrack(t *testing.T) {
	mixed := MediaInfo{Audio: []AudioTrack{
		{Index: 1, Codec: "eac3", Language: "hin", Default: true},
		{Index: 2, Codec: "aac", Language: "eng"},
	}}
	if got := SummaryAudioCodec(mixed); got != "aac" {
		t.Errorf("SummaryAudioCodec = %q, want aac — the English track is what would play", got)
	}

	if got := SummaryAudioCodec(MediaInfo{}); got != "" {
		t.Errorf("SummaryAudioCodec of a silent release = %q, want empty", got)
	}

	// An untagged single track is the common case for a plain release, and it
	// must summarise rather than being skipped as "not the right language".
	untagged := MediaInfo{Audio: []AudioTrack{{Index: 1, Codec: "aac"}}}
	if got := SummaryAudioCodec(untagged); got != "aac" {
		t.Errorf("SummaryAudioCodec of an untagged track = %q, want aac", got)
	}
}

// TestEncodeOmitsNothingLoadBearing guards the zero values that carry meaning.
// Stream index 0 is a real index, and `omitempty` on it would silently move
// every video stream to index 0 on the way back — which happens to be right for
// most files and wrong for the ones that matter.
func TestEncodeOmitsNothingLoadBearing(t *testing.T) {
	info := MediaInfo{
		Video: VideoTrack{Index: 0, Codec: "h264"},
		Audio: []AudioTrack{{Index: 0, Codec: "aac"}},
	}
	raw, err := Encode(info)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	got, ok := Decode(raw)
	if !ok {
		t.Fatal("Decode rejected its own output")
	}
	if got.Video.Index != 0 || len(got.Audio) != 1 || got.Audio[0].Index != 0 {
		t.Errorf("zero indexes did not survive: %+v", got)
	}
	// An SDR release must come back SDR rather than absent-and-therefore-unknown.
	if got.Video.HDRFormat != "" {
		t.Errorf("HDRFormat = %q, want empty", got.Video.HDRFormat)
	}
}
