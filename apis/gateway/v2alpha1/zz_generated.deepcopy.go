//go:build !ignore_autogenerated

/*
Copyright 2022.

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

// Code generated by controller-gen. DO NOT EDIT.

package v2alpha1

import (
	"k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *APIRule) DeepCopyInto(out *APIRule) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new APIRule.
func (in *APIRule) DeepCopy() *APIRule {
	if in == nil {
		return nil
	}
	out := new(APIRule)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *APIRule) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *APIRuleList) DeepCopyInto(out *APIRuleList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]APIRule, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new APIRuleList.
func (in *APIRuleList) DeepCopy() *APIRuleList {
	if in == nil {
		return nil
	}
	out := new(APIRuleList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *APIRuleList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *APIRuleSpec) DeepCopyInto(out *APIRuleSpec) {
	*out = *in
	if in.Hosts != nil {
		in, out := &in.Hosts, &out.Hosts
		*out = make([]*Host, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(Host)
				**out = **in
			}
		}
	}
	if in.Service != nil {
		in, out := &in.Service, &out.Service
		*out = new(Service)
		(*in).DeepCopyInto(*out)
	}
	if in.Gateway != nil {
		in, out := &in.Gateway, &out.Gateway
		*out = new(string)
		**out = **in
	}
	if in.CorsPolicy != nil {
		in, out := &in.CorsPolicy, &out.CorsPolicy
		*out = new(CorsPolicy)
		(*in).DeepCopyInto(*out)
	}
	if in.Rules != nil {
		in, out := &in.Rules, &out.Rules
		*out = make([]Rule, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.Timeout != nil {
		in, out := &in.Timeout, &out.Timeout
		*out = new(Timeout)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new APIRuleSpec.
func (in *APIRuleSpec) DeepCopy() *APIRuleSpec {
	if in == nil {
		return nil
	}
	out := new(APIRuleSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *APIRuleStatus) DeepCopyInto(out *APIRuleStatus) {
	*out = *in
	in.LastProcessedTime.DeepCopyInto(&out.LastProcessedTime)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new APIRuleStatus.
func (in *APIRuleStatus) DeepCopy() *APIRuleStatus {
	if in == nil {
		return nil
	}
	out := new(APIRuleStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CorsPolicy) DeepCopyInto(out *CorsPolicy) {
	*out = *in
	if in.AllowHeaders != nil {
		in, out := &in.AllowHeaders, &out.AllowHeaders
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.AllowMethods != nil {
		in, out := &in.AllowMethods, &out.AllowMethods
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.AllowOrigins != nil {
		in, out := &in.AllowOrigins, &out.AllowOrigins
		*out = make(StringMatch, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = make(map[string]string, len(*in))
				for key, val := range *in {
					(*out)[key] = val
				}
			}
		}
	}
	if in.AllowCredentials != nil {
		in, out := &in.AllowCredentials, &out.AllowCredentials
		*out = new(bool)
		**out = **in
	}
	if in.ExposeHeaders != nil {
		in, out := &in.ExposeHeaders, &out.ExposeHeaders
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.MaxAge != nil {
		in, out := &in.MaxAge, &out.MaxAge
		*out = new(uint64)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CorsPolicy.
func (in *CorsPolicy) DeepCopy() *CorsPolicy {
	if in == nil {
		return nil
	}
	out := new(CorsPolicy)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ExtAuth) DeepCopyInto(out *ExtAuth) {
	*out = *in
	if in.ExternalAuthorizers != nil {
		in, out := &in.ExternalAuthorizers, &out.ExternalAuthorizers
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Restrictions != nil {
		in, out := &in.Restrictions, &out.Restrictions
		*out = new(JwtConfig)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ExtAuth.
func (in *ExtAuth) DeepCopy() *ExtAuth {
	if in == nil {
		return nil
	}
	out := new(ExtAuth)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *JwtAuthentication) DeepCopyInto(out *JwtAuthentication) {
	*out = *in
	if in.FromHeaders != nil {
		in, out := &in.FromHeaders, &out.FromHeaders
		*out = make([]*JwtHeader, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(JwtHeader)
				**out = **in
			}
		}
	}
	if in.FromParams != nil {
		in, out := &in.FromParams, &out.FromParams
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new JwtAuthentication.
func (in *JwtAuthentication) DeepCopy() *JwtAuthentication {
	if in == nil {
		return nil
	}
	out := new(JwtAuthentication)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *JwtAuthorization) DeepCopyInto(out *JwtAuthorization) {
	*out = *in
	if in.RequiredScopes != nil {
		in, out := &in.RequiredScopes, &out.RequiredScopes
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.Audiences != nil {
		in, out := &in.Audiences, &out.Audiences
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new JwtAuthorization.
func (in *JwtAuthorization) DeepCopy() *JwtAuthorization {
	if in == nil {
		return nil
	}
	out := new(JwtAuthorization)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *JwtConfig) DeepCopyInto(out *JwtConfig) {
	*out = *in
	if in.Authentications != nil {
		in, out := &in.Authentications, &out.Authentications
		*out = make([]*JwtAuthentication, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(JwtAuthentication)
				(*in).DeepCopyInto(*out)
			}
		}
	}
	if in.Authorizations != nil {
		in, out := &in.Authorizations, &out.Authorizations
		*out = make([]*JwtAuthorization, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = new(JwtAuthorization)
				(*in).DeepCopyInto(*out)
			}
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new JwtConfig.
func (in *JwtConfig) DeepCopy() *JwtConfig {
	if in == nil {
		return nil
	}
	out := new(JwtConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *JwtConfig) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *JwtHeader) DeepCopyInto(out *JwtHeader) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new JwtHeader.
func (in *JwtHeader) DeepCopy() *JwtHeader {
	if in == nil {
		return nil
	}
	out := new(JwtHeader)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *MutatingWebhook) DeepCopyInto(out *MutatingWebhook) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new MutatingWebhook.
func (in *MutatingWebhook) DeepCopy() *MutatingWebhook {
	if in == nil {
		return nil
	}
	out := new(MutatingWebhook)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Request) DeepCopyInto(out *Request) {
	*out = *in
	if in.Cookies != nil {
		in, out := &in.Cookies, &out.Cookies
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	if in.Headers != nil {
		in, out := &in.Headers, &out.Headers
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Request.
func (in *Request) DeepCopy() *Request {
	if in == nil {
		return nil
	}
	out := new(Request)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Rule) DeepCopyInto(out *Rule) {
	*out = *in
	if in.Service != nil {
		in, out := &in.Service, &out.Service
		*out = new(Service)
		(*in).DeepCopyInto(*out)
	}
	if in.Methods != nil {
		in, out := &in.Methods, &out.Methods
		*out = make([]HttpMethod, len(*in))
		copy(*out, *in)
	}
	if in.NoAuth != nil {
		in, out := &in.NoAuth, &out.NoAuth
		*out = new(bool)
		**out = **in
	}
	if in.Jwt != nil {
		in, out := &in.Jwt, &out.Jwt
		*out = new(JwtConfig)
		(*in).DeepCopyInto(*out)
	}
	if in.ExtAuth != nil {
		in, out := &in.ExtAuth, &out.ExtAuth
		*out = new(ExtAuth)
		(*in).DeepCopyInto(*out)
	}
	if in.Timeout != nil {
		in, out := &in.Timeout, &out.Timeout
		*out = new(Timeout)
		**out = **in
	}
	if in.Request != nil {
		in, out := &in.Request, &out.Request
		*out = new(Request)
		(*in).DeepCopyInto(*out)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Rule.
func (in *Rule) DeepCopy() *Rule {
	if in == nil {
		return nil
	}
	out := new(Rule)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Service) DeepCopyInto(out *Service) {
	*out = *in
	if in.Name != nil {
		in, out := &in.Name, &out.Name
		*out = new(string)
		**out = **in
	}
	if in.Namespace != nil {
		in, out := &in.Namespace, &out.Namespace
		*out = new(string)
		**out = **in
	}
	if in.Port != nil {
		in, out := &in.Port, &out.Port
		*out = new(uint32)
		**out = **in
	}
	if in.IsExternal != nil {
		in, out := &in.IsExternal, &out.IsExternal
		*out = new(bool)
		**out = **in
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Service.
func (in *Service) DeepCopy() *Service {
	if in == nil {
		return nil
	}
	out := new(Service)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in StringMatch) DeepCopyInto(out *StringMatch) {
	{
		in := &in
		*out = make(StringMatch, len(*in))
		for i := range *in {
			if (*in)[i] != nil {
				in, out := &(*in)[i], &(*out)[i]
				*out = make(map[string]string, len(*in))
				for key, val := range *in {
					(*out)[key] = val
				}
			}
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new StringMatch.
func (in StringMatch) DeepCopy() StringMatch {
	if in == nil {
		return nil
	}
	out := new(StringMatch)
	in.DeepCopyInto(out)
	return *out
}
