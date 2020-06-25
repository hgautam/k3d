/*
Copyright © 2020 The k3d Author(s)

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
package load

import (
	"github.com/spf13/cobra"

	"github.com/rancher/k3d/v3/cmd/util"
	"github.com/rancher/k3d/v3/pkg/runtimes"
	"github.com/rancher/k3d/v3/pkg/tools"
	k3d "github.com/rancher/k3d/v3/pkg/types"

	log "github.com/sirupsen/logrus"
)

// NewCmdLoadImage returns a new cobra command
func NewCmdLoadImage() *cobra.Command {

	loadImageOpts := k3d.LoadImageOpts{}

	// create new command
	cmd := &cobra.Command{
		Use:     "image [IMAGE | ARCHIVE [IMAGE | ARCHIVE...]]",
		Short:   "Load an image from docker into a k3d cluster.",
		Long:    `Load an image from docker into a k3d cluster.`,
		Aliases: []string{"images"},
		Args:    cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			images, clusters := parseLoadImageCmd(cmd, args)
			log.Debugf("Load images [%+v] from runtime [%s] into clusters [%+v]", images, runtimes.SelectedRuntime, clusters)
			for _, cluster := range clusters {
				log.Infof("Loading images into '%s'", cluster.Name)
				if err := tools.LoadImagesIntoCluster(cmd.Context(), runtimes.SelectedRuntime, images, &cluster, loadImageOpts); err != nil {
					log.Errorf("Failed to load images into cluster '%s'", cluster.Name)
					log.Errorln(err)
				}
			}
			log.Info("DONE")
		},
	}

	/*********
	 * Flags *
	 *********/
	cmd.Flags().StringArrayP("cluster", "c", []string{k3d.DefaultClusterName}, "Select clusters to load the image to.")
	if err := cmd.RegisterFlagCompletionFunc("cluster", util.ValidArgsAvailableClusters); err != nil {
		log.Fatalln("Failed to register flag completion for '--cluster'", err)
	}

	cmd.Flags().BoolVarP(&loadImageOpts.KeepTar, "keep-tarball", "k", false, "Do not delete the tarball containing the saved images from the shared volume")

	/* Subcommands */

	// done
	return cmd
}

// parseLoadImageCmd parses the command input into variables required to create a cluster
func parseLoadImageCmd(cmd *cobra.Command, args []string) ([]string, []k3d.Cluster) {

	// --cluster
	clusterNames, err := cmd.Flags().GetStringArray("cluster")
	if err != nil {
		log.Fatalln(err)
	}
	clusters := []k3d.Cluster{}
	for _, clusterName := range clusterNames {
		clusters = append(clusters, k3d.Cluster{Name: clusterName})
	}

	// images
	images := args
	if len(images) == 0 {
		log.Fatalln("No images specified!")
	}

	return images, clusters
}
