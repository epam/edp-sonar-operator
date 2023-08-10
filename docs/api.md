# API Reference

Packages:

- [edp.epam.com/v1alpha1](#edpepamcomv1alpha1)
- [v2.edp.epam.com/v1](#v2edpepamcomv1)
- [v2.edp.epam.com/v1alpha1](#v2edpepamcomv1alpha1)

# edp.epam.com/v1alpha1

Resource Types:

- [SonarGroup](#sonargroup)

- [SonarPermissionTemplate](#sonarpermissiontemplate)

- [Sonar](#sonar)




## SonarGroup
<sup><sup>[↩ Parent](#edpepamcomv1alpha1 )</sup></sup>






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
      <td>edp.epam.com/v1alpha1</td>
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
<sup><sup>[↩ Parent](#edpepamcomv1alpha1 )</sup></sup>






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
      <td>edp.epam.com/v1alpha1</td>
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
<sup><sup>[↩ Parent](#edpepamcomv1alpha1 )</sup></sup>






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
      <td>edp.epam.com/v1alpha1</td>
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
        <td><b>secret</b></td>
        <td>string</td>
        <td>
          Secret is the name of the k8s object Secret related to sonar. Secret should contain a user field with a sonar username and a password field with a sonar password. Pass the token in the user field and leave the password field empty for token authentication.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>url</b></td>
        <td>string</td>
        <td>
          Url is the url of sonar instance.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>defaultPermissionTemplate</b></td>
        <td>string</td>
        <td>
          DefaultPermissionTemplate is the name of the default permission template.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#sonarspecsettingsindex">settings</a></b></td>
        <td>[]object</td>
        <td>
          Settings specify which settings should be configured.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Sonar.spec.settings[index]
<sup><sup>[↩ Parent](#sonarspec)</sup></sup>



SonarSetting defines the setting of sonar.

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
        <td><b>key</b></td>
        <td>string</td>
        <td>
          Key is the key of the setting.<br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>fieldValues</b></td>
        <td>map[string]string</td>
        <td>
          Setting field values. To set several values, the parameter must be called once for each value.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>value</b></td>
        <td>string</td>
        <td>
          Value is the value of the setting.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>values</b></td>
        <td>[]string</td>
        <td>
          Setting multi value. To set several values, the parameter must be called once for each value.<br/>
        </td>
        <td>false</td>
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
        <td><b>connected</b></td>
        <td>boolean</td>
        <td>
          Connected shows if operator is connected to sonar.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>error</b></td>
        <td>string</td>
        <td>
          Error represents error message if something went wrong.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>processedSettings</b></td>
        <td>string</td>
        <td>
          ProcessedSettings shows which settings were processed.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>value</b></td>
        <td>string</td>
        <td>
          Value is status of sonar instance. Possible values: GREEN: SonarQube is fully operational YELLOW: SonarQube is usable, but it needs attention in order to be fully operational RED: SonarQube is not operational<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>

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
        <td><b><a href="#sonargroupspec-1">spec</a></b></td>
        <td>object</td>
        <td>
          SonarGroupSpec defines the desired state of SonarGroup.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#sonargroupstatus-1">status</a></b></td>
        <td>object</td>
        <td>
          SonarGroupStatus defines the observed state of SonarGroup.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### SonarGroup.spec
<sup><sup>[↩ Parent](#sonargroup-1)</sup></sup>



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
<sup><sup>[↩ Parent](#sonargroup-1)</sup></sup>



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
        <td><b><a href="#sonarpermissiontemplatespec-1">spec</a></b></td>
        <td>object</td>
        <td>
          SonarPermissionTemplateSpec defines the desired state of SonarPermissionTemplate.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#sonarpermissiontemplatestatus-1">status</a></b></td>
        <td>object</td>
        <td>
          SonarPermissionTemplateStatus defines the observed state of SonarPermissionTemplate.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### SonarPermissionTemplate.spec
<sup><sup>[↩ Parent](#sonarpermissiontemplate-1)</sup></sup>



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
<sup><sup>[↩ Parent](#sonarpermissiontemplate-1)</sup></sup>



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
        <td><b><a href="#sonarspec-1">spec</a></b></td>
        <td>object</td>
        <td>
          SonarSpec defines the desired state of Sonar.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#sonarstatus-1">status</a></b></td>
        <td>object</td>
        <td>
          SonarStatus defines the observed state of Sonar.<br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Sonar.spec
<sup><sup>[↩ Parent](#sonar-1)</sup></sup>



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
        <td><b>secret</b></td>
        <td>string</td>
        <td>
          Secret is the name of the k8s object Secret related to sonar.<br/>
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
        <td><b>exposeToKeycloak</b></td>
        <td>boolean</td>
        <td>
          ExposeToKeycloak specifies whether the sonar should be exposed to keycloak.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#sonarspecgroupsindex">groups</a></b></td>
        <td>[]object</td>
        <td>
          Groups specify which groups should be created.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>plugins</b></td>
        <td>[]string</td>
        <td>
          Plugins specify which plugins should be installed to sonar.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#sonarspecqualitygatesindex">qualityGates</a></b></td>
        <td>[]object</td>
        <td>
          QualityGates specify which quality gates should be created.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#sonarspecsettingsindex-1">settings</a></b></td>
        <td>[]object</td>
        <td>
          Settings specify which settings should be configured.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>url</b></td>
        <td>string</td>
        <td>
          Url is used to explicitly specify the url of sonar. It may not be needed if the sonar is deployed in the same cluster.<br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b><a href="#sonarspecusersindex">users</a></b></td>
        <td>[]object</td>
        <td>
          Users specify which users should be created.<br/>
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


### Sonar.spec.groups[index]
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
        <td><b>name</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>permissions</b></td>
        <td>[]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Sonar.spec.qualityGates[index]
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
        <td><b><a href="#sonarspecqualitygatesindexconditionsindex">conditions</a></b></td>
        <td>[]object</td>
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
        <td><b>setAsDefault</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Sonar.spec.qualityGates[index].conditions[index]
<sup><sup>[↩ Parent](#sonarspecqualitygatesindex)</sup></sup>





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
        <td><b>error</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>metric</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>op</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>period</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Sonar.spec.settings[index]
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
        <td><b>key</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>value</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>valueType</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr></tbody>
</table>


### Sonar.spec.users[index]
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
        <td><b>login</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>username</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>true</td>
      </tr><tr>
        <td><b>exposeToJenkins</b></td>
        <td>boolean</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>group</b></td>
        <td>string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr><tr>
        <td><b>permissions</b></td>
        <td>[]string</td>
        <td>
          <br/>
        </td>
        <td>false</td>
      </tr></tbody>
</table>


### Sonar.status
<sup><sup>[↩ Parent](#sonar-1)</sup></sup>



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