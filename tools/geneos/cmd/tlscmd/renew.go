/*
Copyright © 2022 ITRS Group

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/

package tlscmd

import (
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	_ "embed"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/spf13/cobra"

	"github.com/itrs-group/cordial/pkg/config"
	"github.com/itrs-group/cordial/tools/geneos/cmd"
	"github.com/itrs-group/cordial/tools/geneos/internal/geneos"
	"github.com/itrs-group/cordial/tools/geneos/internal/instance"
)

func init() {
	tlsCmd.AddCommand(renewCmd)
}

//go:embed _docs/renew.md
var renewCmdDescription string

var renewCmd = &cobra.Command{
	Use:          "renew [TYPE] [NAME...]",
	Short:        "Renew instance certificates",
	Long:         renewCmdDescription,
	SilenceUsage: true,
	Annotations: map[string]string{
		cmd.AnnotationWildcard:  "true",
		cmd.AnnotationNeedsHome: "true",
	},
	Run: func(command *cobra.Command, _ []string) {
		ct, names := cmd.TypeNames(command)
		instance.Do(geneos.GetHost(cmd.Hostname), ct, names, renewInstanceCert).Write(os.Stdout)
	},
}

// renew an instance certificate, use private key if it exists
func renewInstanceCert(c geneos.Instance, _ ...any) (resp *instance.Response) {
	resp = instance.NewResponse(c)

	hostname, _ := os.Hostname()
	if !c.Host().IsLocal() {
		hostname = c.Host().GetString("hostname")
	}

	serial, err := rand.Prime(rand.Reader, 64)
	if err != nil {
		return
	}
	expires := time.Now().AddDate(1, 0, 0).Truncate(24 * time.Hour)
	template := x509.Certificate{
		SerialNumber: serial,
		Subject: pkix.Name{
			CommonName: fmt.Sprintf("geneos %s %s", c.Type(), c.Name()),
		},
		NotBefore:      time.Now().Add(-60 * time.Second),
		NotAfter:       expires,
		KeyUsage:       x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:    []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
		MaxPathLenZero: true,
		DNSNames:       []string{hostname},
		// IPAddresses:    []net.IP{net.ParseIP("127.0.0.1")},
	}

	rootCert, err := instance.ReadRootCert()
	resp.Err = err
	if resp.Err != nil {
		return
	}

	signingCert, err := instance.ReadSigningCert()
	resp.Err = err
	if resp.Err != nil {
		return
	}
	signingKey, err := config.ReadPrivateKey(geneos.LOCAL, path.Join(config.AppConfigDir(), geneos.SigningCertFile+".key"))
	resp.Err = err
	if resp.Err != nil {
		return
	}

	// read existing key or create a new one
	existingKey, _ := instance.ReadKey(c)
	cert, key, err := config.CreateCertificateAndKey(&template, signingCert, signingKey, existingKey)
	resp.Err = err
	if resp.Err != nil {
		return
	}

	if resp.Err = instance.WriteCert(c, cert); resp.Err != nil {
		return
	}

	if existingKey == nil {
		if resp.Err = instance.WriteKey(c, key); resp.Err != nil {
			return
		}
	}

	chainfile := instance.PathOf(c, "certchain")
	if chainfile == "" {
		chainfile = path.Join(c.Home(), "chain.pem")
		c.Config().Set("certchain", chainfile)
	}

	if resp.Err = config.WriteCertChain(c.Host(), chainfile, signingCert, rootCert); resp.Err != nil {
		return
	}

	if resp.Err = instance.SaveConfig(c); resp.Err != nil {
		return
	}

	resp.Completed = append(resp.Completed, fmt.Sprintf("certificate renewed (expires %s)", expires))
	return
}
