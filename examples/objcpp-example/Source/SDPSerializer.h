// SDPSerializer.h
// Objective-C interface for SDP serialization

#import <Foundation/Foundation.h>
#import "User.h"

NS_ASSUME_NONNULL_BEGIN

@interface SDPSerializer : NSObject

/// Encode a User object to binary data
/// @param user The user to encode
/// @param error Error pointer (optional)
/// @return Encoded NSData or nil on failure
+ (nullable NSData *)encodeUser:(User *)user error:(NSError **)error;

/// Decode binary data to a User object
/// @param data The data to decode
/// @param error Error pointer (optional)
/// @return Decoded User or nil on failure
+ (nullable User *)decodeUser:(NSData *)data error:(NSError **)error;

/// Validate that data is within SDP size limits
+ (BOOL)isValidData:(NSData *)data;

/// Get the size of encoded user (estimate)
+ (NSUInteger)estimatedSizeForUser:(User *)user;

@end

NS_ASSUME_NONNULL_END
