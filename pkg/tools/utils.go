/*

Copyright (C) 2019  Daniele Rondina <geaaru@sabayonlinux.org>
Credits goes also to Gogs authors, some code portions and re-implemented design
are also coming from the Gogs project, which is using the go-macaron framework
and was really source of ispiration. Kudos to them!

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.

*/
package tools

import (
	"fmt"
	"os"
	"time"
)

type TimeSorter []time.Time

func (t TimeSorter) Len() int           { return len(t) }
func (t TimeSorter) Swap(i, j int)      { t[i], t[j] = t[j], t[i] }
func (t TimeSorter) Less(i, j int) bool { return t[i].Before(t[j]) }

func MkdirIfNotExist(dir string, mode os.FileMode) (*os.FileInfo, error) {
	var err error
	var info os.FileInfo

	if dir == "" {
		return nil, fmt.Errorf("Invalid directory")
	}

	if info, err = os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, mode)
		if err != nil {
			return nil, err
		}
	}

	return &info, nil
}
