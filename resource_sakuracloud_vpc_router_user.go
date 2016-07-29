package sakuracloud

import (
	"fmt"

	"bytes"
	"github.com/hashicorp/terraform/helper/hashcode"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/yamamoto-febc/libsacloud/api"
	"github.com/yamamoto-febc/libsacloud/sacloud"
)

func resourceSakuraCloudVPCRouterRemoteAccessUser() *schema.Resource {
	return &schema.Resource{
		Create: resourceSakuraCloudVPCRouterRemoteAccessUserCreate,
		Read:   resourceSakuraCloudVPCRouterRemoteAccessUserRead,
		Delete: resourceSakuraCloudVPCRouterRemoteAccessUserDelete,
		Schema: map[string]*schema.Schema{
			"vpc_router_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"password": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"zone": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ForceNew:     true,
				Description:  "target SakuraCloud zone",
				ValidateFunc: validateStringInWord([]string{"is1a", "is1b", "tk1a", "tk1v"}),
			},
		},
	}
}

func resourceSakuraCloudVPCRouterRemoteAccessUserCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*api.Client)
	zone, ok := d.GetOk("zone")
	if ok {
		client.Zone = zone.(string)
	}

	routerID := d.Get("vpc_router_id").(string)
	sakuraMutexKV.Lock(routerID)
	defer sakuraMutexKV.Unlock(routerID)

	vpcRouter, err := client.VPCRouter.Read(routerID)
	if err != nil {
		return fmt.Errorf("Couldn't find SakuraCloud VPCRouter resource: %s", err)
	}

	remoteAccessUser := expandVPCRouterRemoteAccessUser(d)
	if vpcRouter.Settings == nil {
		vpcRouter.InitVPCRouterSetting()
	}

	vpcRouter.Settings.Router.AddRemoteAccessUser(remoteAccessUser.UserName, remoteAccessUser.Password)
	vpcRouter, err = client.VPCRouter.UpdateSetting(routerID, vpcRouter)
	if err != nil {
		return fmt.Errorf("Failed to enable SakuraCloud VPCRouterRemoteAccessUser resource: %s", err)
	}
	_, err = client.VPCRouter.Config(routerID)
	if err != nil {
		return fmt.Errorf("Couldn'd apply SakuraCloud VPCRouter config: %s", err)
	}

	d.SetId(vpcRouterRemoteAccessUserIDHash(routerID, remoteAccessUser))
	return resourceSakuraCloudVPCRouterRemoteAccessUserRead(d, meta)
}

func resourceSakuraCloudVPCRouterRemoteAccessUserRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*api.Client)
	zone, ok := d.GetOk("zone")
	if ok {
		client.Zone = zone.(string)
	}

	routerID := d.Get("vpc_router_id").(string)
	vpcRouter, err := client.VPCRouter.Read(routerID)
	if err != nil {
		return fmt.Errorf("Couldn't find SakuraCloud VPCRouter resource: %s", err)
	}

	remoteAccessUser := expandVPCRouterRemoteAccessUser(d)
	if vpcRouter.Settings != nil && vpcRouter.Settings.Router != nil && vpcRouter.Settings.Router.RemoteAccessUsers != nil &&
		vpcRouter.Settings.Router.FindRemoteAccessUser(remoteAccessUser.UserName, remoteAccessUser.Password) != nil {
		d.Set("name", remoteAccessUser.UserName)
		d.Set("password", remoteAccessUser.Password)
	}

	d.Set("zone", client.Zone)

	return nil
}

func resourceSakuraCloudVPCRouterRemoteAccessUserDelete(d *schema.ResourceData, meta interface{}) error {

	client := meta.(*api.Client)
	zone, ok := d.GetOk("zone")
	if ok {
		client.Zone = zone.(string)
	}

	routerID := d.Get("vpc_router_id").(string)
	sakuraMutexKV.Lock(routerID)
	defer sakuraMutexKV.Unlock(routerID)

	vpcRouter, err := client.VPCRouter.Read(routerID)
	if err != nil {
		return fmt.Errorf("Couldn't find SakuraCloud VPCRouter resource: %s", err)
	}

	if vpcRouter.Settings.Router.RemoteAccessUsers != nil {

		remoteAccessUser := expandVPCRouterRemoteAccessUser(d)
		vpcRouter.Settings.Router.RemoveRemoteAccessUser(remoteAccessUser.UserName, remoteAccessUser.Password)

		vpcRouter, err = client.VPCRouter.UpdateSetting(routerID, vpcRouter)
		if err != nil {
			return fmt.Errorf("Failed to delete SakuraCloud VPCRouterRemoteAccessUser resource: %s", err)
		}

		_, err = client.VPCRouter.Config(routerID)
		if err != nil {
			return fmt.Errorf("Couldn'd apply SakuraCloud VPCRouter config: %s", err)
		}
	}

	d.SetId("")
	return nil
}

func vpcRouterRemoteAccessUserIDHash(routerID string, s *sacloud.VPCRouterRemoteAccessUsersConfig) string {
	var buf bytes.Buffer
	buf.WriteString(fmt.Sprintf("%s-", routerID))
	buf.WriteString(fmt.Sprintf("%s-", s.UserName))
	buf.WriteString(fmt.Sprintf("%s", s.Password))

	return fmt.Sprintf("%d", hashcode.String(buf.String()))
}

func expandVPCRouterRemoteAccessUser(d *schema.ResourceData) *sacloud.VPCRouterRemoteAccessUsersConfig {

	var remoteAccessUser = &sacloud.VPCRouterRemoteAccessUsersConfig{
		UserName: d.Get("name").(string),
		Password: d.Get("password").(string),
	}

	return remoteAccessUser
}
