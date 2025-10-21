// User.m
// Model class implementation

#import "User.h"

@implementation UserMetadata
@end

@implementation User

- (instancetype)init {
    if (self = [super init]) {
        _tags = @[];
    }
    return self;
}

- (NSString *)description {
    return [NSString stringWithFormat:@"<User id=%lu username=%@ email=%@ age=%lu active=%@ tags=%lu metadata=%@>",
            (unsigned long)_userId,
            _username,
            _email,
            (unsigned long)_age,
            _isActive ? @"YES" : @"NO",
            (unsigned long)_tags.count,
            _metadata ? @"present" : @"nil"];
}

@end
