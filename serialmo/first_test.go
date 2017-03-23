// file serialmo/first_test.go

package serialmo  // import "github.com/bstarynk/monimelt/serialmo"

import (
	"fmt"
	"testing"
)

func TestSerialToString(t *testing.T) {
	s1 := SerialMo(2734358116516558954) // _3fZo81e6aIa
	fmt.Printf("TestSerialToString s1=%d\n", s1)
	s1s := s1.ToString()
	fmt.Printf("s1:%s\n", s1s)
	if s1s != "_3fZo81e6aIa" {
		t.Errorf("TestSerialToString bad s1s='%s'", s1s)
	}

}

func TestFromStringSerial(t *testing.T) {
	const s2s = "_4Fgo2LZq1AS" /// 3915796129876347282
	const s2n = 3915796129876347282
	fmt.Printf("TestFromStringSerial s2s=%s s2n=%d=%#x\n", s2s, s2n, s2n)
	s2, e := FromString(s2s)
	fmt.Printf("s2=%d=%#x e=%v\n", s2, s2, e)
	s2str := s2.ToString()
	fmt.Printf("s2:%s\n", s2.ToString())
	if s2str != s2s {
		t.Errorf("TestFromStringSerial fail s2s='%s' s2str='%s' s2=%#x",
			s2s, s2str, s2)
	}
}

func TestFirst(t *testing.T) {
	s1, e := FromUint64(4096)
	fmt.Printf("TestFirst s1=%d=%#x e=%v\n", s1, s1, e)
	s1s := s1.ToString()
	fmt.Printf("s1s='%s'\n", s1s)
	s1n, e := FromString(s1s)
	fmt.Printf("s1n=%d=%#x e=%v\n", s1n, s1n, e)
	///
	s2, e := FromUint64(62 * 62 * 62)
	fmt.Printf("s2=%d=%#x e=%v\n", s2, s2, e)
	s2s := s2.ToString()
	fmt.Printf("s2s='%s'\n", s2s)
	s2n, e := FromString(s2s)
	fmt.Printf("s2n=%d=%#x e=%v\n", s2n, s2n, e)
	//
	const s3str = "_0000000A000"
	const s3nn = 10 * 62 * 62 * 62
	s3, e := FromString(s3str)
	fmt.Printf("s3=%d=%#x s3str='%s' s3nn=%d=%#x e=%v\n", s3, s3, s3str,
		s3nn, s3nn, e)
	s3s := s3.ToString()
	fmt.Printf("s3s='%s'\n", s3s)
	sr := RandomSerial()
	srs := sr.ToString()
	fmt.Printf("sr=%d=%#x='%s' bucket#%d\n", sr, sr, srs, sr.BucketNum())
	srf, e := FromString(srs)
	if srf != sr {
		t.Errorf("TestFirst failure sr=%d=%#x='%s' e=%v\n",
			sr, sr, srs, e)
	}
	sr = RandomSerial()
	srs = sr.ToString()
	fmt.Printf("sr=%d=%#x='%s' bucket#%d\n", sr, sr, srs, sr.BucketNum())
	if FromCheckedString(srs) != sr {
		t.Errorf("TestFirst failure sr=%d=%#x='%s'\n", sr, sr, srs)
	}
	sr = RandomSerial()
	srs = sr.ToString()
	fmt.Printf("sr=%d=%#x='%s' bucket#%d\n", sr, sr, srs, sr.BucketNum())
	sb := RandomOfBucket(sr.BucketNum())
	sbs := sb.ToString()
	fmt.Printf("sb=%d=%#x='%s' bucket#%d\n", sb, sb, sbs, sb.BucketNum())
	if FromCheckedString(srs) != sr {
		t.Errorf("TestFirst failure sr=%d=%#x='%s'\n", sr, sr, srs)
	}
	if sb.BucketNum() != sr.BucketNum() {
		t.Errorf("TestFirst failure sr=%d=%#x='%s' bucket#%d non isobucket with sb=%d=%#x='%s' bucket#%d\n",
			sr, sr, srs, sr.BucketNum(),
			sb, sb, sbs, sb.BucketNum())
	}

}
