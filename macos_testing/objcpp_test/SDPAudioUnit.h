//
//  SDPAudioUnit.h
//  Objective-C wrapper for C++ SDP AudioUnit implementation
//
//  This header provides an Objective-C interface to the C++ encoder/decoder.
//  Can be imported in pure Objective-C or Swift (pre-Swift 6).
//

#import <Foundation/Foundation.h>

NS_ASSUME_NONNULL_BEGIN

// Forward declarations matching C++ types
@interface SDPParameter : NSObject
@property (nonatomic, assign) uint64_t address;
@property (nonatomic, strong) NSString *displayName;
@property (nonatomic, strong) NSString *identifier;
@property (nonatomic, strong) NSString *unit;
@property (nonatomic, assign) float minValue;
@property (nonatomic, assign) float maxValue;
@property (nonatomic, assign) float defaultValue;
@property (nonatomic, assign) float currentValue;
@property (nonatomic, assign) uint32_t rawFlags;
@property (nonatomic, assign) BOOL isWritable;
@property (nonatomic, assign) BOOL canRamp;
@end

@interface SDPPlugin : NSObject
@property (nonatomic, strong) NSString *name;
@property (nonatomic, strong) NSString *manufacturerID;
@property (nonatomic, strong) NSString *componentType;
@property (nonatomic, strong) NSString *componentSubtype;
@property (nonatomic, strong) NSArray<SDPParameter *> *parameters;
@end

@interface SDPPluginRegistry : NSObject
@property (nonatomic, strong) NSArray<SDPPlugin *> *plugins;
@property (nonatomic, assign) uint32_t totalPluginCount;
@property (nonatomic, assign) uint32_t totalParameterCount;
@end

// Encoder/Decoder interface
@interface SDPAudioUnitCodec : NSObject

// Decode binary SDP data to Objective-C objects
+ (nullable SDPPluginRegistry *)decodePluginRegistry:(NSData *)data
                                               error:(NSError **)error;

// Encode Objective-C objects to binary SDP data
+ (nullable NSData *)encodePluginRegistry:(SDPPluginRegistry *)registry
                                    error:(NSError **)error;

// Get encoded size (for pre-allocating buffers)
+ (size_t)pluginRegistrySize:(SDPPluginRegistry *)registry;

@end

NS_ASSUME_NONNULL_END
