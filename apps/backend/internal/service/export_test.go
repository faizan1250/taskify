package service

import "context"

// DeleteS3ObjectWithRetryForTest exposes deleteS3ObjectWithRetry to the
// external test package. The deleter argument only needs to satisfy the
// DeleteObject method set; callers do not name the interface type.
func DeleteS3ObjectWithRetryForTest(
	ctx context.Context,
	deleter s3ObjectDeleter,
	bucket string,
	key string,
) (int, error) {
	return deleteS3ObjectWithRetry(ctx, deleter, bucket, key)
}
