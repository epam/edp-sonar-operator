//go:build !ignore_autogenerated
// +build !ignore_autogenerated

// Code generated by controller-gen. DO NOT EDIT.

package v1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *EdpSpec) DeepCopyInto(out *EdpSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new EdpSpec.
func (in *EdpSpec) DeepCopy() *EdpSpec {
	if in == nil {
		return nil
	}
	out := new(EdpSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *GroupPermission) DeepCopyInto(out *GroupPermission) {
	*out = *in
	if in.Permissions != nil {
		in, out := &in.Permissions, &out.Permissions
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new GroupPermission.
func (in *GroupPermission) DeepCopy() *GroupPermission {
	if in == nil {
		return nil
	}
	out := new(GroupPermission)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Sonar) DeepCopyInto(out *Sonar) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Sonar.
func (in *Sonar) DeepCopy() *Sonar {
	if in == nil {
		return nil
	}
	out := new(Sonar)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Sonar) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SonarGroup) DeepCopyInto(out *SonarGroup) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SonarGroup.
func (in *SonarGroup) DeepCopy() *SonarGroup {
	if in == nil {
		return nil
	}
	out := new(SonarGroup)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *SonarGroup) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SonarGroupList) DeepCopyInto(out *SonarGroupList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]SonarGroup, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SonarGroupList.
func (in *SonarGroupList) DeepCopy() *SonarGroupList {
	if in == nil {
		return nil
	}
	out := new(SonarGroupList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *SonarGroupList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SonarGroupSpec) DeepCopyInto(out *SonarGroupSpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SonarGroupSpec.
func (in *SonarGroupSpec) DeepCopy() *SonarGroupSpec {
	if in == nil {
		return nil
	}
	out := new(SonarGroupSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SonarGroupStatus) DeepCopyInto(out *SonarGroupStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SonarGroupStatus.
func (in *SonarGroupStatus) DeepCopy() *SonarGroupStatus {
	if in == nil {
		return nil
	}
	out := new(SonarGroupStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SonarList) DeepCopyInto(out *SonarList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Sonar, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SonarList.
func (in *SonarList) DeepCopy() *SonarList {
	if in == nil {
		return nil
	}
	out := new(SonarList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *SonarList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SonarPermissionTemplate) DeepCopyInto(out *SonarPermissionTemplate) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SonarPermissionTemplate.
func (in *SonarPermissionTemplate) DeepCopy() *SonarPermissionTemplate {
	if in == nil {
		return nil
	}
	out := new(SonarPermissionTemplate)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *SonarPermissionTemplate) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SonarPermissionTemplateList) DeepCopyInto(out *SonarPermissionTemplateList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]SonarPermissionTemplate, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SonarPermissionTemplateList.
func (in *SonarPermissionTemplateList) DeepCopy() *SonarPermissionTemplateList {
	if in == nil {
		return nil
	}
	out := new(SonarPermissionTemplateList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *SonarPermissionTemplateList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SonarPermissionTemplateSpec) DeepCopyInto(out *SonarPermissionTemplateSpec) {
	*out = *in
	if in.GroupPermissions != nil {
		in, out := &in.GroupPermissions, &out.GroupPermissions
		*out = make([]GroupPermission, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SonarPermissionTemplateSpec.
func (in *SonarPermissionTemplateSpec) DeepCopy() *SonarPermissionTemplateSpec {
	if in == nil {
		return nil
	}
	out := new(SonarPermissionTemplateSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SonarPermissionTemplateStatus) DeepCopyInto(out *SonarPermissionTemplateStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SonarPermissionTemplateStatus.
func (in *SonarPermissionTemplateStatus) DeepCopy() *SonarPermissionTemplateStatus {
	if in == nil {
		return nil
	}
	out := new(SonarPermissionTemplateStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SonarSpec) DeepCopyInto(out *SonarSpec) {
	*out = *in
	out.EdpSpec = in.EdpSpec
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SonarSpec.
func (in *SonarSpec) DeepCopy() *SonarSpec {
	if in == nil {
		return nil
	}
	out := new(SonarSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SonarStatus) DeepCopyInto(out *SonarStatus) {
	*out = *in
	in.LastTimeUpdated.DeepCopyInto(&out.LastTimeUpdated)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SonarStatus.
func (in *SonarStatus) DeepCopy() *SonarStatus {
	if in == nil {
		return nil
	}
	out := new(SonarStatus)
	in.DeepCopyInto(out)
	return out
}