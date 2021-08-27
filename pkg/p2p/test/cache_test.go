/*
   Copyright The Accelerated Container Image Authors

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package test

import (
	"math/rand"
	"runtime"
	"testing"

	"github.com/alibaba/accelerated-container-image/pkg/p2p/cache"
	"github.com/alibaba/accelerated-container-image/pkg/p2p/fs"
	"github.com/alibaba/accelerated-container-image/pkg/p2p/util"

	"github.com/stretchr/testify/assert"
)

func testCacheGetOrRefillHelper(t *testing.T, config *cache.Config) {
	t.Helper()
	Assert := assert.New(t)
	c := cache.NewCachePool(config)
	for i := 0; i < 100; i++ {
		fileName := util.GetRandomString(10)
		fileContent := []byte(getData())
		for j := 0; j < 10; j++ {
			for seg := range fs.NewRangeSplit(0, 128*1024, int64(len(fileContent)), int64(len(fileContent))).AllParts() {
				wg.Add(1)
				go func(offset int64, size int) {
					defer wg.Done()
					res, err := c.GetOrRefill(fileName, offset, size, func() ([]byte, error) {
						return fileContent[offset : offset+int64(size)], nil
					})
					if Assert.Equal(nil, err) && Assert.Equal(size, len(res)) {
						expected := fileContent[offset : offset+int64(size)]
						checkLen := fs.Min(100, size)
						Assert.Equal(expected[:checkLen], res[:checkLen])
						Assert.Equal(expected[len(res)-checkLen:], res[len(res)-checkLen:])
					}
				}(seg.Index, seg.Count)
			}
		}
	}
	wg.Wait()
	runtime.GC()
}

func TestCacheGetOrRefill(t *testing.T) {
	testCacheGetOrRefillHelper(t, &cache.Config{CacheSize: 100 * 1024 * 1024, MaxEntry: 0, CacheMedia: media})
	testCacheGetOrRefillHelper(t, &cache.Config{CacheSize: 90 * 1024 * 1024, MaxEntry: 0, CacheMedia: media})
	testCacheGetOrRefillHelper(t, &cache.Config{CacheSize: 10 * 1024 * 1024, MaxEntry: 0, CacheMedia: media})
	testCacheGetOrRefillHelper(t, &cache.Config{CacheSize: 1 * 1024 * 1024, MaxEntry: 0, CacheMedia: media})
	testCacheGetOrRefillHelper(t, &cache.Config{CacheSize: 0, MaxEntry: 0, CacheMedia: media})
}

func testCacheGetPutHostHelper(t *testing.T, config *cache.Config) {
	t.Helper()
	Assert := assert.New(t)
	c := cache.NewCachePool(config)
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			filename := util.GetRandomString(10)
			host := util.GetRandomString(1024)
			// get
			res, hit := c.GetHost(filename)
			Assert.Equal(false, hit)
			Assert.Equal("", res)
			// put
			hit = c.PutHost(filename, host)
			Assert.Equal(true, hit)
			// get
			res, hit = c.GetHost(filename)
			Assert.Equal(true, hit)
			Assert.Equal(host, res)
			// del
			c.DelHost(filename)
			// get
			res, hit = c.GetHost(filename)
			Assert.Equal(false, hit)
			Assert.Equal("", res)
		}()
	}
	wg.Wait()
}

func TestCacheGetPutHost(t *testing.T) {
	testCacheGetPutHostHelper(t, &cache.Config{CacheSize: 0, MaxEntry: 1000 * 1024 * 1024, CacheMedia: media})
}

func testCacheGetPutLengthHelper(t *testing.T, config *cache.Config) {
	t.Helper()
	Assert := assert.New(t)
	c := cache.NewCachePool(config)
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			filename := util.GetRandomString(10)
			length := rand.Int63n(1024)
			// get
			res, hit := c.GetLen(filename)
			Assert.Equal(false, hit)
			Assert.Equal(int64(0), res)
			// put
			hit = c.PutLen(filename, length)
			Assert.Equal(true, hit)
			// get
			res, hit = c.GetLen(filename)
			Assert.Equal(true, hit)
			Assert.Equal(length, res)
		}()
	}
	wg.Wait()
}

func TestCacheGetPutLength(t *testing.T) {
	testCacheGetPutLengthHelper(t, &cache.Config{CacheSize: 0, MaxEntry: 1000 * 1024 * 1024, CacheMedia: media})
}
