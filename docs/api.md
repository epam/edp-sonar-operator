# API Reference

Packages:

- [v2.edp.epam.com/v1](#v2edpepamcomv1)
- [v2.edp.epam.com/v1alpha1](#v2edpepamcomv1alpha1)

# v2.edp.epam.com/v1

Resource Types:

- [SonarGroup](#sonargroup)

- [SonarPermissionTemplate](#sonarpermissiontemplate)

- [Sonar](#sonar)




## SonarGroup
<sup><sup>[↩ Parent](#v2edpepamcomv1 )</sup></sup>






SonarGroup is the Schema for the sonar group API.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
      <td><b>apiVersion</b></td>
      <td>string</td>
      <td>v2.edp.epam.com/v1</td>
      <td>true</td>
      </tr>
      <tr>
      <td><b>kind</b></td>
      <td>string</td>
      <td>SonarGroup</td>
      <td>true</td>
      </tr>
      <tr>
      <td><b><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.20/#objectmeta-v1-meta">metadata</a></b></td>
      <td>object</td>
      <td>Refer to the Kubernetes API documentation for the fields of the `metadata` field.</td>
      <td>true</td>
      </tr><tr>
        <td><b><a href="#sonargroupspec">spec</a></b></td>
        <td>object</td>
        <td>
          SonarGroupSpec defines the desired state of SonarGroup.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#sonargroupstatus">status</a></b></td>
        <td>object</td>
        <td>
          SonarGroupStatus defines the observed state of SonarGroup.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### SonarGroup.spec
<sup><sup>[↩ Parent](#sonargroup)</sup></sup>



SonarGroupSpec defines the desired state of SonarGroup.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name is a group name.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>sonarOwner</b></td>
        <td>string</td>
        <td>
          SonarOwner is a name of root sonar custom resource.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>description</b></td>
        <td>string</td>
        <td>
          Description of sonar group.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### SonarGroup.status
<sup><sup>[↩ Parent](#sonargroup)</sup></sup>



SonarGroupStatus defines the observed state of SonarGroup.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>failureCount</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int64<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>id</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>value</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>

## SonarPermissionTemplate
<sup><sup>[↩ Parent](#v2edpepamcomv1 )</sup></sup>






SonarPermissionTemplate is the Schema for the sonar permission template API.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
      <td><b>apiVersion</b></td>
      <td>string</td>
      <td>v2.edp.epam.com/v1</td>
      <td>true</td>
      </tr>
      <tr>
      <td><b>kind</b></td>
      <td>string</td>
      <td>SonarPermissionTemplate</td>
      <td>true</td>
      </tr>
      <tr>
      <td><b><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.20/#objectmeta-v1-meta">metadata</a></b></td>
      <td>object</td>
      <td>Refer to the Kubernetes API documentation for the fields of the `metadata` field.</td>
      <td>true</td>
      </tr><tr>
        <td><b><a href="#sonarpermissiontemplatespec">spec</a></b></td>
        <td>object</td>
        <td>
          SonarPermissionTemplateSpec defines the desired state of SonarPermissionTemplate.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#sonarpermissiontemplatestatus">status</a></b></td>
        <td>object</td>
        <td>
          SonarPermissionTemplateStatus defines the observed state of SonarPermissionTemplate.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### SonarPermissionTemplate.spec
<sup><sup>[↩ Parent](#sonarpermissiontemplate)</sup></sup>



SonarPermissionTemplateSpec defines the desired state of SonarPermissionTemplate.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name is a group name.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>projectKeyPattern</b></td>
        <td>string</td>
        <td>
          ProjectKeyPattern is key pattern. Must be a valid Java regular expression.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>sonarOwner</b></td>
        <td>string</td>
        <td>
          SonarOwner is a name of root sonar custom resource.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>description</b></td>
        <td>string</td>
        <td>
          Description of sonar permission template.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#sonarpermissiontemplatespecgrouppermissionsindex">groupPermissions</a></b></td>
        <td>[]object</td>
        <td>
          GroupPermissions adds a group to a permission template.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### SonarPermissionTemplate.spec.groupPermissions[index]
<sup><sup>[↩ Parent](#sonarpermissiontemplatespec)</sup></sup>



GroupPermission represents the group and its permissions.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>groupName</b></td>
        <td>string</td>
        <td>
          Group name or 'anyone' (case insensitive). Example value sonar-administrators.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>permissions</b></td>
        <td>[]string</td>
        <td>
          Permissions is a list of permissions. Possible values: admin, codeviewer, issueadmin, securityhotspotadmin, scan, user.<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### SonarPermissionTemplate.status
<sup><sup>[↩ Parent](#sonarpermissiontemplate)</sup></sup>



SonarPermissionTemplateStatus defines the observed state of SonarPermissionTemplate.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>failureCount</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int64<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>id</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>value</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>

## Sonar
<sup><sup>[↩ Parent](#v2edpepamcomv1 )</sup></sup>






Sonar is the Schema for the sonars API.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
      <td><b>apiVersion</b></td>
      <td>string</td>
      <td>v2.edp.epam.com/v1</td>
      <td>true</td>
      </tr>
      <tr>
      <td><b>kind</b></td>
      <td>string</td>
      <td>Sonar</td>
      <td>true</td>
      </tr>
      <tr>
      <td><b><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.20/#objectmeta-v1-meta">metadata</a></b></td>
      <td>object</td>
      <td>Refer to the Kubernetes API documentation for the fields of the `metadata` field.</td>
      <td>true</td>
      </tr><tr>
        <td><b><a href="#sonarspec">spec</a></b></td>
        <td>object</td>
        <td>
          SonarSpec defines the desired state of Sonar.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#sonarstatus">status</a></b></td>
        <td>object</td>
        <td>
          SonarStatus defines the observed state of Sonar.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Sonar.spec
<sup><sup>[↩ Parent](#sonar)</sup></sup>



SonarSpec defines the desired state of Sonar.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b><a href="#sonarspecedpspec">edpSpec</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>basePath</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>defaultPermissionTemplate</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Sonar.spec.edpSpec
<sup><sup>[↩ Parent](#sonarspec)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>dnsWildcard</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### Sonar.status
<sup><sup>[↩ Parent](#sonar)</sup></sup>



SonarStatus defines the observed state of Sonar.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>available</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>externalUrl</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>lastTimeUpdated</b></td>
        <td>string</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: date-time<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>status</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>

# v2.edp.epam.com/v1alpha1

Resource Types:

- [SonarGroup](#sonargroup)

- [SonarPermissionTemplate](#sonarpermissiontemplate)

- [Sonar](#sonar)




## SonarGroup
<sup><sup>[↩ Parent](#v2edpepamcomv1alpha1 )</sup></sup>






SonarGroup is the Schema for the sonar group API.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
      <td><b>apiVersion</b></td>
      <td>string</td>
      <td>v2.edp.epam.com/v1alpha1</td>
      <td>true</td>
      </tr>
      <tr>
      <td><b>kind</b></td>
      <td>string</td>
      <td>SonarGroup</td>
      <td>true</td>
      </tr>
      <tr>
      <td><b><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.20/#objectmeta-v1-meta">metadata</a></b></td>
      <td>object</td>
      <td>Refer to the Kubernetes API documentation for the fields of the `metadata` field.</td>
      <td>true</td>
      </tr><tr>
        <td><b><a href="#sonargroupspec-1">spec</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#sonargroupstatus-1">status</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### SonarGroup.spec
<sup><sup>[↩ Parent](#sonargroup-1)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name is a group name.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>sonarOwner</b></td>
        <td>string</td>
        <td>
          SonarOwner is a name of root sonar custom resource.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>description</b></td>
        <td>string</td>
        <td>
          Description of sonar group.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### SonarGroup.status
<sup><sup>[↩ Parent](#sonargroup-1)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>failureCount</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int64<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>id</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>value</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>

## SonarPermissionTemplate
<sup><sup>[↩ Parent](#v2edpepamcomv1alpha1 )</sup></sup>






SonarPermissionTemplate is the Schema for the sonar permission template API.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
      <td><b>apiVersion</b></td>
      <td>string</td>
      <td>v2.edp.epam.com/v1alpha1</td>
      <td>true</td>
      </tr>
      <tr>
      <td><b>kind</b></td>
      <td>string</td>
      <td>SonarPermissionTemplate</td>
      <td>true</td>
      </tr>
      <tr>
      <td><b><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.20/#objectmeta-v1-meta">metadata</a></b></td>
      <td>object</td>
      <td>Refer to the Kubernetes API documentation for the fields of the `metadata` field.</td>
      <td>true</td>
      </tr><tr>
        <td><b><a href="#sonarpermissiontemplatespec-1">spec</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#sonarpermissiontemplatestatus-1">status</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### SonarPermissionTemplate.spec
<sup><sup>[↩ Parent](#sonarpermissiontemplate-1)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name is a group name.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>projectKeyPattern</b></td>
        <td>string</td>
        <td>
          ProjectKeyPattern is key pattern. Must be a valid Java regular expression.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>sonarOwner</b></td>
        <td>string</td>
        <td>
          SonarOwner is a name of root sonar custom resource.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>description</b></td>
        <td>string</td>
        <td>
          Description of sonar permission template.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#sonarpermissiontemplatespecgrouppermissionsindex-1">groupPermissions</a></b></td>
        <td>[]object</td>
        <td>
          GroupPermissions adds a group to a permission template.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### SonarPermissionTemplate.spec.groupPermissions[index]
<sup><sup>[↩ Parent](#sonarpermissiontemplatespec-1)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>groupName</b></td>
        <td>string</td>
        <td>
          Group name or 'anyone' (case insensitive). Example value sonar-administrators.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>permissions</b></td>
        <td>[]string</td>
        <td>
          Permissions is a list of permissions. Possible values: admin, codeviewer, issueadmin, securityhotspotadmin, scan, user.<br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### SonarPermissionTemplate.status
<sup><sup>[↩ Parent](#sonarpermissiontemplate-1)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>failureCount</b></td>
        <td>integer</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: int64<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>id</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>value</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>

## Sonar
<sup><sup>[↩ Parent](#v2edpepamcomv1alpha1 )</sup></sup>






Sonar is the Schema for the sonars API.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
      <td><b>apiVersion</b></td>
      <td>string</td>
      <td>v2.edp.epam.com/v1alpha1</td>
      <td>true</td>
      </tr>
      <tr>
      <td><b>kind</b></td>
      <td>string</td>
      <td>Sonar</td>
      <td>true</td>
      </tr>
      <tr>
      <td><b><a href="https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.20/#objectmeta-v1-meta">metadata</a></b></td>
      <td>object</td>
      <td>Refer to the Kubernetes API documentation for the fields of the `metadata` field.</td>
      <td>true</td>
      </tr><tr>
        <td><b><a href="#sonarspec-1">spec</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#sonarstatus-1">status</a></b></td>
        <td>object</td>
        <td>
          SonarStatus defines the observed state of Sonar<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Sonar.spec
<sup><sup>[↩ Parent](#sonar-1)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>dbImage</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b><a href="#sonarspecedpspec-1">edpSpec</a></b></td>
        <td>object</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>image</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>initImage</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>version</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>basePath</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>defaultPermissionTemplate</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#sonarspecimagepullsecretsindex">imagePullSecrets</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#sonarspecvolumesindex">volumes</a></b></td>
        <td>[]object</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Sonar.spec.edpSpec
<sup><sup>[↩ Parent](#sonarspec-1)</sup></sup>





<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>dnsWildcard</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### Sonar.spec.imagePullSecrets[index]
<sup><sup>[↩ Parent](#sonarspec-1)</sup></sup>



LocalObjectReference contains enough information to let you locate the referenced object inside the same namespace.

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names TODO: Add other useful fields. apiVersion, kind, uid?<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Sonar.spec.volumes[index]
<sup><sup>[↩ Parent](#sonarspec-1)</sup></sup>



SonarSpec defines the desired state of Sonar

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>capacity</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>name</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>storage_class</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Sonar.status
<sup><sup>[↩ Parent](#sonar-1)</sup></sup>



SonarStatus defines the observed state of Sonar

<table>
    <thead>
        <tr>
            <th>Name</th>
            <th>Type</th>
            <th>Description</th>
            <th>Required</th>
        </tr>
    </thead>
    <tbody><tr>
        <td><b>available</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>externalUrl</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>lastTimeUpdated</b></td>
        <td>string</td>
        <td>
          <br/>
          <br/>
            <i>Format</i>: date-time<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>status</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>