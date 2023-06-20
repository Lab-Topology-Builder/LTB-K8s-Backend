package controllers

import (
	"context"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	ltbv1alpha1 "github.com/Lab-Topology-Builder/LTB-K8s-Backend/api/v1alpha1"
	util "github.com/Lab-Topology-Builder/LTB-K8s-Backend/util"
)

type LabTemplateReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=ltb-backend.ltb,resources=labtemplates,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=ltb-backend.ltb,resources=labtemplates/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=ltb-backend.ltb,resources=labtemplates/finalizers,verbs=update
//+kubebuilder:rbac:groups=ltb-backend.ltb,resources=nodetypes,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=ltb-backend.ltb,resources=nodetypes/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=ltb-backend.ltb,resources=nodetypes/finalizers,verbs=update

func (r *LabTemplateReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	l := log.FromContext(ctx)
	labTemplate := &ltbv1alpha1.LabTemplate{}
	l.Info("Reconciling LabTemplate")
	err := r.Get(ctx, req.NamespacedName, labTemplate)
	if err != nil {
		l.Error(err, "Failed to get labtemplate, ignoring must have been deleted")
		return ctrl.Result{Requeue: true, RequeueAfter: 2 * time.Second}, client.IgnoreNotFound(err)
	}
	nodes := &labTemplate.Spec.Nodes
	for i := 0; i < len(*nodes); i++ {
		nodetype := &ltbv1alpha1.NodeType{}
		err := r.Get(ctx, client.ObjectKey{Namespace: labTemplate.Namespace, Name: (*nodes)[i].NodeTypeRef.Type}, nodetype)
		if err != nil {
			l.Error(err, "Failed to get nodetype")
			return ctrl.Result{}, client.IgnoreNotFound(err)
		}
		var renderedNodeSpec strings.Builder
		if err = util.ParseAndRenderTemplate(nodetype, &renderedNodeSpec, (*nodes)[i]); err != nil {
			l.Error(err, "Failed to render template")
			return ctrl.Result{}, err
		}
		(*nodes)[i].RenderedNodeSpec = renderedNodeSpec.String()
	}

	err = r.Update(ctx, labTemplate)
	if err != nil {
		l.Error(err, "Failed to update labtemplate")
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (r *LabTemplateReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&ltbv1alpha1.LabTemplate{}).
		Complete(r)
}
