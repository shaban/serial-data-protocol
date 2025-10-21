// User.h
// Model class representing a user

#import <Foundation/Foundation.h>

NS_ASSUME_NONNULL_BEGIN

@interface UserMetadata : NSObject

@property (nonatomic, assign) uint64_t createdAt;
@property (nonatomic, assign) uint64_t updatedAt;
@property (nonatomic, assign) NSUInteger loginCount;

@end

@interface User : NSObject

@property (nonatomic, assign) NSUInteger userId;
@property (nonatomic, strong) NSString *username;
@property (nonatomic, strong) NSString *email;
@property (nonatomic, assign) NSUInteger age;
@property (nonatomic, assign) BOOL isActive;
@property (nonatomic, strong) NSArray<NSString *> *tags;
@property (nonatomic, strong, nullable) UserMetadata *metadata;

@end

NS_ASSUME_NONNULL_END
