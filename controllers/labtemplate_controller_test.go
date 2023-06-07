/*
Copyright 2023 Jan Untersander, Tsigereda Nebai Kidane.

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

package controllers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	ctrl "sigs.k8s.io/controller-runtime"
)

func TestLabTemplateReconciler_Reconcile(t *testing.T) {
	type args struct {
		ctx context.Context
		req ctrl.Request
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    ctrl.Result
		want1   error
		wantErr bool
	}{
		// {
		// 	name: "Empty request",
		// 	args: args{
		// 		ctx: context.Background(),
		// 		req: ctrl.Request{},
		// 	},
		// 	want:    ctrl.Result{},
		// 	want1:   nil,
		// 	wantErr: false,
		// },
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &LabTemplateReconciler{
				Client: tt.fields.Client,
				Scheme: tt.fields.Scheme,
			}
			got, err := tr.Reconcile(tt.args.ctx, tt.args.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("LabTemplateReconciler.Reconcile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.want1, err)
		})
	}
}

func TestLabTemplateReconciler_SetupWithManager(t *testing.T) {
	type args struct {
		mgr ctrl.Manager
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &LabTemplateReconciler{
				Client: tt.fields.Client,
				Scheme: tt.fields.Scheme,
			}
			if err := r.SetupWithManager(tt.args.mgr); (err != nil) != tt.wantErr {
				t.Errorf("LabTemplateReconciler.SetupWithManager() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
