//
//  SDPAudioUnit.h
//  Zero-copy Objective-C wrapper for C++ SDP AudioUnit
//
//  This header provides a thin Objective-C interface that keeps data in C++.
//  No object allocation for decoding - just wraps the C++ struct.
//

#import <Foundation/Foundation.h>

NS_ASSUME_NONNULL_BEGIN

// Opaque pointer to C++ PluginRegistry
// Data stays in C++, only exposed via methods
@interface SDPPluginRegistry : NSObject

// Decode from NSData (keeps data in C++)
+ (nullable instancetype)decodeFromData:(NSData *)data error:(NSError **)error;

// Encode to NSData
- (nullable NSData *)encodeWithError:(NSError **)error;

// Read-only properties (no object allocation)
@property (nonatomic, readonly) uint32_t totalPluginCount;
@property (nonatomic, readonly) uint32_t totalParameterCount;
@property (nonatomic, readonly) NSInteger pluginCount;

// Access plugins by index (returns C++ reference wrapped in lightweight object)
- (NSString *)pluginNameAtIndex:(NSInteger)index;
- (NSInteger)parameterCountForPluginAtIndex:(NSInteger)index;

// Access parameters by plugin index and parameter index
- (NSString *)parameterDisplayNameForPlugin:(NSInteger)pluginIndex parameter:(NSInteger)paramIndex;
- (uint64_t)parameterAddressForPlugin:(NSInteger)pluginIndex parameter:(NSInteger)paramIndex;
- (float)parameterCurrentValueForPlugin:(NSInteger)pluginIndex parameter:(NSInteger)paramIndex;

// Size calculation
- (size_t)encodedSize;

@end

NS_ASSUME_NONNULL_END
