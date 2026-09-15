// Letters and digits that can't be confused when read aloud or copied from a screen: no 0/o, 1/l/i.
const ALPHABET = 'abcdefghjkmnpqrstuvwxyz23456789'

/** A random password students can type from a board: 8 characters, about 40 bits. */
export function generatePassword(length = 8, random: (bytes: Uint8Array) => Uint8Array = (b) => crypto.getRandomValues(b)): string {
  // Bytes at or above the largest multiple of the alphabet size are skipped, so every character is equally likely.
  const limit = 256 - (256 % ALPHABET.length)
  let result = ''
  while (result.length < length) {
    for (const byte of random(new Uint8Array(length * 2))) {
      if (byte < limit && result.length < length) {
        result += ALPHABET[byte % ALPHABET.length]
      }
    }
  }
  return result
}
