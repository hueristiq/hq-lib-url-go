package tld

// Pseudo is a sorted list of widely used unofficial or pseudo top-level domains (TLDs).
//
// These domains are not part of the official IANA TLD registry but serve specific purposes
// in private networks, special-use cases, or decentralized domain name systems.
//
// References:
//   - Wikipedia: https://en.wikipedia.org/wiki/Pseudo-top-level_domain
//   - IETF Draft: https://tools.ietf.org/html/draft-grothoff-iesg-special-use-p2p-names-00
//   - IANA Special-Use Domains: https://www.iana.org/assignments/special-use-domain-names/special-use-domain-names.xhtml
var Pseudo = []string{
	`bit`,       // Namecoin - a decentralized domain system based on the Namecoin blockchain.
	`example`,   // Example domain - reserved for use in documentation and examples.
	`exit`,      // Tor exit node - used for identifying Tor exit nodes in the Tor network.
	`gnu`,       // GNS by public key - GNU Name System, a decentralized name system.
	`i2p`,       // I2P network - Invisible Internet Project, an anonymous network layer.
	`invalid`,   // Invalid domain - reserved for invalid domain names.
	`local`,     // Local network - used in local networking environments.
	`localhost`, // Local network - refers to the local loopback interface (127.0.0.1).
	`test`,      // Test domain - reserved for use in testing environments.
	`zkey`,      // GNS domain name - used in the GNU Name System for public-key based domain names.
}
