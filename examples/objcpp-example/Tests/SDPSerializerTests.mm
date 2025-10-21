// SDPSerializerTests.mm
// Unit tests for SDP serializer

#import <XCTest/XCTest.h>
#import "SDPSerializer.h"
#import "User.h"

@interface SDPSerializerTests : XCTestCase
@end

@implementation SDPSerializerTests

- (User *)createTestUser {
    User *user = [[User alloc] init];
    user.userId = 12345;
    user.username = @"testuser";
    user.email = @"test@example.com";
    user.age = 25;
    user.isActive = YES;
    user.tags = @[@"developer", @"macos", @"ios"];
    
    UserMetadata *metadata = [[UserMetadata alloc] init];
    metadata.createdAt = 1634567890;
    metadata.updatedAt = 1634567900;
    metadata.loginCount = 42;
    user.metadata = metadata;
    
    return user;
}

- (void)testBasicEncodeDecode {
    // Create test user
    User *original = [self createTestUser];
    
    // Encode
    NSError *encodeError = nil;
    NSData *encoded = [SDPSerializer encodeUser:original error:&encodeError];
    
    XCTAssertNotNil(encoded, @"Encoding should succeed");
    XCTAssertNil(encodeError, @"Should not have encoding error");
    XCTAssertGreaterThan(encoded.length, 0, @"Encoded data should not be empty");
    
    NSLog(@"Encoded size: %lu bytes", (unsigned long)encoded.length);
    
    // Decode
    NSError *decodeError = nil;
    User *decoded = [SDPSerializer decodeUser:encoded error:&decodeError];
    
    XCTAssertNotNil(decoded, @"Decoding should succeed");
    XCTAssertNil(decodeError, @"Should not have decoding error");
    
    // Verify all fields
    XCTAssertEqual(decoded.userId, original.userId);
    XCTAssertEqualObjects(decoded.username, original.username);
    XCTAssertEqualObjects(decoded.email, original.email);
    XCTAssertEqual(decoded.age, original.age);
    XCTAssertEqual(decoded.isActive, original.isActive);
    XCTAssertEqual(decoded.tags.count, original.tags.count);
    
    for (NSUInteger i = 0; i < original.tags.count; i++) {
        XCTAssertEqualObjects(decoded.tags[i], original.tags[i]);
    }
    
    XCTAssertNotNil(decoded.metadata);
    XCTAssertEqual(decoded.metadata.createdAt, original.metadata.createdAt);
    XCTAssertEqual(decoded.metadata.updatedAt, original.metadata.updatedAt);
    XCTAssertEqual(decoded.metadata.loginCount, original.metadata.loginCount);
}

- (void)testUserWithoutMetadata {
    User *user = [[User alloc] init];
    user.userId = 999;
    user.username = @"nometa";
    user.email = @"nometa@test.com";
    user.age = 30;
    user.isActive = NO;
    user.tags = @[@"test"];
    user.metadata = nil; // No metadata
    
    NSData *encoded = [SDPSerializer encodeUser:user error:nil];
    XCTAssertNotNil(encoded);
    
    User *decoded = [SDPSerializer decodeUser:encoded error:nil];
    XCTAssertNotNil(decoded);
    XCTAssertNil(decoded.metadata, @"Metadata should be nil");
    XCTAssertEqual(decoded.userId, user.userId);
}

- (void)testEmptyArrays {
    User *user = [[User alloc] init];
    user.userId = 100;
    user.username = @"empty";
    user.email = @"empty@test.com";
    user.age = 20;
    user.isActive = YES;
    user.tags = @[]; // Empty array
    
    NSData *encoded = [SDPSerializer encodeUser:user error:nil];
    XCTAssertNotNil(encoded);
    
    User *decoded = [SDPSerializer decodeUser:encoded error:nil];
    XCTAssertNotNil(decoded);
    XCTAssertEqual(decoded.tags.count, 0);
}

- (void)testInvalidData {
    NSData *invalidData = [@"this is not valid SDP data" dataUsingEncoding:NSUTF8StringEncoding];
    
    NSError *error = nil;
    User *decoded = [SDPSerializer decodeUser:invalidData error:&error];
    
    XCTAssertNil(decoded, @"Should fail to decode invalid data");
    XCTAssertNotNil(error, @"Should return an error");
    XCTAssertEqualObjects(error.domain, @"SDPErrorDomain");
    
    NSLog(@"Expected error: %@", error.localizedDescription);
}

- (void)testNilInput {
    NSError *encodeError = nil;
    NSData *encoded = [SDPSerializer encodeUser:nil error:&encodeError];
    XCTAssertNil(encoded);
    XCTAssertNotNil(encodeError);
    
    NSError *decodeError = nil;
    User *decoded = [SDPSerializer decodeUser:nil error:&decodeError];
    XCTAssertNil(decoded);
    XCTAssertNotNil(decodeError);
}

- (void)testDataValidation {
    NSData *smallData = [NSData dataWithBytes:"test" length:4];
    XCTAssertTrue([SDPSerializer isValidData:smallData]);
    
    // Create data that's too large (> 128 MB)
    NSMutableData *hugeData = [NSMutableData dataWithLength:129 * 1024 * 1024];
    XCTAssertFalse([SDPSerializer isValidData:hugeData]);
}

- (void)testSizeEstimation {
    User *user = [self createTestUser];
    NSUInteger estimated = [SDPSerializer estimatedSizeForUser:user];
    
    NSData *encoded = [SDPSerializer encodeUser:user error:nil];
    NSUInteger actual = encoded.length;
    
    NSLog(@"Estimated size: %lu, Actual size: %lu", 
          (unsigned long)estimated, (unsigned long)actual);
    
    // Estimate should be close (within 20% typically)
    XCTAssertLessThanOrEqual(actual, estimated + 50);
    XCTAssertGreaterThanOrEqual(actual, estimated - 50);
}

- (void)testPerformance {
    User *user = [self createTestUser];
    
    // Measure encoding performance
    [self measureBlock:^{
        for (int i = 0; i < 1000; i++) {
            @autoreleasepool {
                NSData *encoded = [SDPSerializer encodeUser:user error:nil];
                (void)encoded; // Suppress unused warning
            }
        }
    }];
    
    NSLog(@"Measured encoding performance: 1000 iterations");
}

- (void)testRoundtripPerformance {
    User *user = [self createTestUser];
    NSData *encoded = [SDPSerializer encodeUser:user error:nil];
    
    // Measure decode performance
    [self measureBlock:^{
        for (int i = 0; i < 1000; i++) {
            @autoreleasepool {
                User *decoded = [SDPSerializer decodeUser:encoded error:nil];
                (void)decoded;
            }
        }
    }];
    
    NSLog(@"Measured decoding performance: 1000 iterations");
}

- (void)testLargeStrings {
    User *user = [[User alloc] init];
    user.userId = 999;
    
    // Create a large username (1 MB)
    NSMutableString *largeString = [NSMutableString stringWithCapacity:1024 * 1024];
    for (int i = 0; i < 1024 * 1024; i++) {
        [largeString appendString:@"a"];
    }
    
    user.username = largeString;
    user.email = @"test@example.com";
    user.age = 25;
    user.isActive = YES;
    user.tags = @[];
    
    NSData *encoded = [SDPSerializer encodeUser:user error:nil];
    XCTAssertNotNil(encoded);
    
    User *decoded = [SDPSerializer decodeUser:encoded error:nil];
    XCTAssertNotNil(decoded);
    XCTAssertEqual(decoded.username.length, user.username.length);
}

- (void)testManyTags {
    User *user = [[User alloc] init];
    user.userId = 555;
    user.username = @"tagmaster";
    user.email = @"tags@test.com";
    user.age = 28;
    user.isActive = YES;
    
    // Create array with many tags
    NSMutableArray *tags = [NSMutableArray arrayWithCapacity:1000];
    for (int i = 0; i < 1000; i++) {
        [tags addObject:[NSString stringWithFormat:@"tag_%d", i]];
    }
    user.tags = tags;
    
    NSData *encoded = [SDPSerializer encodeUser:user error:nil];
    XCTAssertNotNil(encoded);
    
    User *decoded = [SDPSerializer decodeUser:encoded error:nil];
    XCTAssertNotNil(decoded);
    XCTAssertEqual(decoded.tags.count, 1000);
}

@end
