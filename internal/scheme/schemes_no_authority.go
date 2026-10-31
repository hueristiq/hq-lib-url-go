package scheme

// NoAuthority is a sorted list of well-known URL schemes that do not require an authority component.
//
// These schemes use a colon (":") instead of "://" and do not require a host or authority part.
// This format is common for protocols that address specific resources or services directly.
//
// The list includes both officially registered schemes and commonly used unofficial schemes.
var NoAuthority = []string{
	`bitcoin`, // Bitcoin - Used to create Bitcoin payment URIs.
	`cid`,     // Content-ID - Identifies a specific piece of content.
	`data`,    // Data - Embeds data inline as if it were an external resource.
	`file`,    // Files - Refers to files on the local filesystem.
	`magnet`,  // Torrent magnets - Used for referencing torrents without relying on a tracker.
	`mailto`,  // Mail - Refers to email addresses for sending mail.
	`mid`,     // Message-ID - Refers to email message IDs.
	`sms`,     // SMS - Used for sending SMS messages.
	`tel`,     // Telephone - Refers to telephone numbers.
	`xmpp`,    // XMPP - Used for addressing XMPP (Jabber) communication services.
}
