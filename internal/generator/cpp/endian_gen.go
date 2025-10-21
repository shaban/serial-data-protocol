package cpp

// GenerateEndianHeader generates portable endianness conversion macros
func GenerateEndianHeader() string {
	return `/* sdp_endian.h - Portable endianness conversion for SDP wire format
 * 
 * SDP wire format is ALWAYS little-endian.
 * This header provides macros to convert host byte order to/from little-endian.
 * 
 * Supported platforms:
 * - Linux (glibc)
 * - macOS
 * - Windows (assumes little-endian Intel/AMD)
 * - BSD systems
 * - Generic fallback with manual byte swapping
 */

#ifndef SDP_ENDIAN_H
#define SDP_ENDIAN_H

#include <stdint.h>
#include <string.h>

/* Detect platform and define conversion macros */
#if defined(__linux__) || defined(__CYGWIN__)
    /* Linux - use endian.h */
    #include <endian.h>
    #define SDP_HTOLE16(x) htole16(x)
    #define SDP_HTOLE32(x) htole32(x)
    #define SDP_HTOLE64(x) htole64(x)
    #define SDP_LE16TOH(x) le16toh(x)
    #define SDP_LE32TOH(x) le32toh(x)
    #define SDP_LE64TOH(x) le64toh(x)

#elif defined(__APPLE__)
    /* macOS - use libkern/OSByteOrder.h */
    #include <libkern/OSByteOrder.h>
    #define SDP_HTOLE16(x) OSSwapHostToLittleInt16(x)
    #define SDP_HTOLE32(x) OSSwapHostToLittleInt32(x)
    #define SDP_HTOLE64(x) OSSwapHostToLittleInt64(x)
    #define SDP_LE16TOH(x) OSSwapLittleToHostInt16(x)
    #define SDP_LE32TOH(x) OSSwapLittleToHostInt32(x)
    #define SDP_LE64TOH(x) OSSwapLittleToHostInt64(x)

#elif defined(__OpenBSD__) || defined(__NetBSD__) || defined(__FreeBSD__) || defined(__DragonFly__)
    /* BSD systems - use sys/endian.h */
    #include <sys/endian.h>
    #define SDP_HTOLE16(x) htole16(x)
    #define SDP_HTOLE32(x) htole32(x)
    #define SDP_HTOLE64(x) htole64(x)
    #define SDP_LE16TOH(x) le16toh(x)
    #define SDP_LE32TOH(x) le32toh(x)
    #define SDP_LE64TOH(x) le64toh(x)

#elif defined(_WIN32) || defined(_WIN64)
    /* Windows - assume little-endian (x86/x64) */
    /* On Windows, Intel/AMD CPUs are little-endian, so no conversion needed */
    #define SDP_HTOLE16(x) (x)
    #define SDP_HTOLE32(x) (x)
    #define SDP_HTOLE64(x) (x)
    #define SDP_LE16TOH(x) (x)
    #define SDP_LE32TOH(x) (x)
    #define SDP_LE64TOH(x) (x)

#else
    /* Generic fallback - manual byte swapping */
    
    /* Byte swap functions */
    static inline uint16_t sdp_bswap16(uint16_t x) {
        return (x >> 8) | (x << 8);
    }
    
    static inline uint32_t sdp_bswap32(uint32_t x) {
        return ((x & 0xff000000u) >> 24) |
               ((x & 0x00ff0000u) >>  8) |
               ((x & 0x0000ff00u) <<  8) |
               ((x & 0x000000ffu) << 24);
    }
    
    static inline uint64_t sdp_bswap64(uint64_t x) {
        return ((x & 0xff00000000000000ull) >> 56) |
               ((x & 0x00ff000000000000ull) >> 40) |
               ((x & 0x0000ff0000000000ull) >> 24) |
               ((x & 0x000000ff00000000ull) >>  8) |
               ((x & 0x00000000ff000000ull) <<  8) |
               ((x & 0x0000000000ff0000ull) << 24) |
               ((x & 0x000000000000ff00ull) << 40) |
               ((x & 0x00000000000000ffull) << 56);
    }
    
    /* Detect endianness at compile time if possible */
    #if defined(__BYTE_ORDER__) && defined(__ORDER_LITTLE_ENDIAN__) && defined(__ORDER_BIG_ENDIAN__)
        #if __BYTE_ORDER__ == __ORDER_LITTLE_ENDIAN__
            /* Little-endian system - no conversion needed */
            #define SDP_HTOLE16(x) (x)
            #define SDP_HTOLE32(x) (x)
            #define SDP_HTOLE64(x) (x)
            #define SDP_LE16TOH(x) (x)
            #define SDP_LE32TOH(x) (x)
            #define SDP_LE64TOH(x) (x)
        #elif __BYTE_ORDER__ == __ORDER_BIG_ENDIAN__
            /* Big-endian system - byte swap required */
            #define SDP_HTOLE16(x) sdp_bswap16(x)
            #define SDP_HTOLE32(x) sdp_bswap32(x)
            #define SDP_HTOLE64(x) sdp_bswap64(x)
            #define SDP_LE16TOH(x) sdp_bswap16(x)
            #define SDP_LE32TOH(x) sdp_bswap32(x)
            #define SDP_LE64TOH(x) sdp_bswap64(x)
        #else
            #error "Unknown byte order"
        #endif
    #else
        /* Runtime detection (slower, but works everywhere) */
        static inline int sdp_is_little_endian(void) {
            const uint16_t test = 1;
            return *(const uint8_t*)&test == 1;
        }
        
        #define SDP_HTOLE16(x) (sdp_is_little_endian() ? (x) : sdp_bswap16(x))
        #define SDP_HTOLE32(x) (sdp_is_little_endian() ? (x) : sdp_bswap32(x))
        #define SDP_HTOLE64(x) (sdp_is_little_endian() ? (x) : sdp_bswap64(x))
        #define SDP_LE16TOH(x) (sdp_is_little_endian() ? (x) : sdp_bswap16(x))
        #define SDP_LE32TOH(x) (sdp_is_little_endian() ? (x) : sdp_bswap32(x))
        #define SDP_LE64TOH(x) (sdp_is_little_endian() ? (x) : sdp_bswap64(x))
    #endif
#endif

/* Float/double conversion helpers */
static inline uint32_t sdp_f32_to_le(float f) {
    uint32_t u;
    memcpy(&u, &f, sizeof(f));
    return SDP_HTOLE32(u);
}

static inline float sdp_le_to_f32(uint32_t u) {
    uint32_t le = SDP_LE32TOH(u);
    float f;
    memcpy(&f, &le, sizeof(f));
    return f;
}

static inline uint64_t sdp_f64_to_le(double d) {
    uint64_t u;
    memcpy(&u, &d, sizeof(d));
    return SDP_HTOLE64(u);
}

static inline double sdp_le_to_f64(uint64_t u) {
    uint64_t le = SDP_LE64TOH(u);
    double d;
    memcpy(&d, &le, sizeof(d));
    return d;
}

#endif /* SDP_ENDIAN_H */
`
}
