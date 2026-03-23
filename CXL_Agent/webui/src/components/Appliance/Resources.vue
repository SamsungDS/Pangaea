<!-- Copyright (c) 2024 Seagate Technology LLC and/or its Affiliates -->
<template>
  <v-container style="padding: 0">
    <!-- Active filter banner (set by clicking ResourceId in Memory table) -->
    <div v-if="activeResourceIdFilter" class="pa-2">
      <v-chip label density="comfortable" class="ma-1">
        Filter: {{ activeResourceIdFilter }}
        <v-icon class="ml-2" size="small" @click.stop="clearResourceFilter">mdi-close</v-icon>
      </v-chip>
    </div>

    <v-data-table
      :headers="headers"
      fixed-header
      height="310"
      :items="selectedBladeResources"
    >
      <template v-slot:[`item.compositionStatus.compositionState`]="{ value }">
        <v-chip :color="getStatusColor(value)">
          {{ value }}
        </v-chip>
      </template>

      <!-- ✅ Actions column (same UX as Memory.vue) -->
      <template v-slot:[`item.actions`]="{ item }">
        <!-- Unused / Composed / Shared: Assign action (port list filtered to USP ports; excludes current USP port) -->
        <template v-if="isUnused(item) || isComposed(item) || isShared(item)">
          <v-icon
            size="small"
            :disabled="isComposed(item) || isShared(item) ? !findMemoryRegionByResourceId(item.id) : false"
            @click="openAssignDialogByStatus(item)"
          >
            mdi-link-plus
          </v-icon>

          <!-- Composed / Shared: Free action (single action; backend handles unassign+free) -->
          <template v-if="isComposed(item) || isShared(item)">
            <span class="ml-2"></span>
            <v-icon
              size="small"
              :disabled="!findMemoryRegionByResourceId(item.id)"
              @click="freeMemoryFromResource(item)"
            >
              mdi-delete
            </v-icon>
          </template>

          <v-tooltip activator="parent" location="end">
            <template v-if="isUnused(item)">
              Assign this Unused Resource to an Upstream Port (USP ports only).
            </template>
            <template v-else>
              Assign this Resource to another Upstream Port (USP ports only). Free will unassign (if needed) and free the memory.
            </template>
          </v-tooltip>
        </template>

        <!-- No action available -->
        <template v-else>
          <v-icon size="small" disabled>mdi-link-plus</v-icon>
          <span class="ml-2"></span>
          <v-icon size="small" disabled>mdi-delete</v-icon>
          <v-tooltip activator="parent" location="end">
            Actions are available only when Status is Unused / Composed / Shared.
          </v-tooltip>
        </template>
      </template>
    </v-data-table>

    <!-- Assign (Resource -> select Upstream Port) -->
    <v-dialog v-model="dialogAssignUnusedResource" max-width="600px">
      <v-card>
        <v-alert
          color="warning"
          icon="$warning"
          title="Alert"
          variant="tonal"
        >
          Binding unused resources to upstream ports may take some time at the switch level.
        </v-alert>
        <v-divider></v-divider>

        <v-card-text>
          <div>
            ResourceId: <strong>{{ selectedUnusedResource?.id }}</strong>
          </div>
          <v-autocomplete
            v-model="assignUnusedResourcePort"
            label="Assign to Port"
            :items="assignableUspPortsForSelectedResource"
            id="inputAssignUnusedResourcePort"
          ></v-autocomplete>
        </v-card-text>

        <v-divider></v-divider>
        <v-card-action>
          <v-spacer></v-spacer>
          <div class="text-end">
            <v-btn
              color="yellow-darken-4"
              variant="text"
              @click="dialogAssignUnusedResource = false"
              id="cancelAssignUnusedResource"
            >cancel</v-btn>
            <v-btn
              color="#1428A0"
              variant="text"
              :disabled="!assignUnusedResourcePort"
              @click="assignUnusedResourceConfirm"
              id="confirmAssignUnusedResource"
            >Assign Resource</v-btn>
          </div>
        </v-card-action>
      </v-card>
    </v-dialog>

    <v-dialog v-model="waitAssignUnusedResource">
      <v-row align-content="center" class="fill-height" justify="center">
        <v-col cols="6">
          <v-progress-linear color="#1428A0" height="50" indeterminate rounded>
            <template v-slot:default>
              <div class="text-center">Assigning Resource, please wait...</div>
            </template>
          </v-progress-linear>
        </v-col>
      </v-row>
    </v-dialog>

    <v-dialog v-model="assignUnusedResourceSuccess" max-width="600px">
      <v-sheet
        elevation="12"
        max-width="600"
        rounded="lg"
        width="100%"
        class="pa-4 text-center mx-auto"
      >
        <v-icon class="mb-5" color="success" icon="mdi-check-circle" size="112"></v-icon>
        <h2 class="text-h5 mb-6">Assign resource succeeded!</h2>
        <p class="mb-4 text-medium-emphasis text-body-2">
          ResourceId: <br />{{ selectedUnusedResource?.id }}
          <br />Port: <br />{{ assignedUnusedResourcePortSnapshot }}
          <br />New Memory ID: <br />{{ newComposedMemoryId || "(unknown)" }}
        </p>
        <v-divider class="mb-4"></v-divider>
        <div class="text-end">
          <v-btn
            class="text-none"
            color="success"
            rounded
            variant="flat"
            width="90"
            id="assignUnusedResourceSuccess"
            @click="assignUnusedResourceSuccess = false"
          >Done</v-btn>
        </div>
      </v-sheet>
    </v-dialog>

    <v-dialog v-model="assignUnusedResourceFailure" max-width="600px">
      <v-sheet
        elevation="12"
        max-width="600"
        rounded="lg"
        width="100%"
        class="pa-4 text-center mx-auto"
      >
        <v-icon class="mb-5" color="error" icon="mdi-alert-circle" size="112"></v-icon>
        <h2 class="text-h5 mb-6">Assign resource failed!</h2>
        <p class="mb-4 text-medium-emphasis text-body-2">
          {{ assignUnusedResourceError }}
        </p>
        <v-divider class="mb-4"></v-divider>
        <div class="text-end">
          <v-btn
            class="text-none"
            color="error"
            rounded
            variant="flat"
            width="90"
            id="assignUnusedResourceFailure"
            @click="assignUnusedResourceFailure = false"
          >Done</v-btn>
        </div>
      </v-sheet>
    </v-dialog>

    <!-- ========================= -->
    <!-- Below dialogs are copied from Memory.vue (as-is) -->
    <!-- ========================= -->

    <v-dialog v-model="dialogAssignUnassign" max-width="600px">
      <v-card>
        <v-alert color="warning" icon="$warning" title="Alert" variant="tonal">
          Due to limited protections, the CXL-Host <strong> MUST </strong> be
          powered down when being
          <strong> {{ this.operation }}ed </strong> memory.
        </v-alert>
        <v-divider></v-divider>
        <v-card-text>
          <div v-if="assign">
            To assign <strong>{{ this.selectedMemoryRegion?.id }}</strong> to a
            port, please <strong> power down </strong> the to be connected
            CXL-Host and select the port from the dropdown, then click the green
            button.
          </div>
          <div v-else>
            To unassign
            <strong>{{ this.selectedMemoryRegion?.id }}</strong> from
            <strong>{{ this.selectedMemoryRegion?.memoryAppliancePort }}</strong
            >, please click the green button after
            <strong> powering down </strong> the connected CXL-Host.
          </div>
        </v-card-text>

        <v-autocomplete
          v-if="assign"
          v-model="assignPort"
          id="inputSelectedPort"
          label="Assign to Port"
          :items="PortIdArray"
        ></v-autocomplete>

        <v-divider></v-divider>
        <v-card-action>
          <v-spacer></v-spacer>
          <div class="text-end">
            <v-btn
              color="yellow-darken-4"
              variant="text"
              @click="dialogAssignUnassign = false"
              id="cancelAssignOrUnassign"
            >cancel</v-btn
            >
            <v-btn
              color="#1428A0"
              variant="text"
              @click="assignUnassignPort(this.operation)"
              id="confirmAssignOrUnassign"
            >{{ assign ? "Assign Memory" : "UnAssign Memroy" }}</v-btn
            >
          </div>
        </v-card-action>
      </v-card>
    </v-dialog>

    <v-dialog v-model="dialogNoPortForAssignMemory" max-width="600px">
      <v-sheet
        elevation="12"
        max-width="600"
        rounded="lg"
        width="100%"
        class="pa-4 text-center mx-auto"
      >
        <v-icon
          class="mb-5"
          color="error"
          icon="mdi-alert-circle"
          size="112"
        ></v-icon>
        <h2 class="text-h5 mb-6">No available ports</h2>
        <p class="mb-4 text-medium-emphasis text-body-2">
          All ports are assigned, please unassign one.
        </p>
        <v-divider class="mb-4"></v-divider>
        <div class="text-end">
          <v-btn
            class="text-none"
            color="error"
            rounded
            variant="flat"
            width="90"
            id="noPortForAssignMemory"
            @click="dialogNoPortForAssignMemory = false"
          >
            Done
          </v-btn>
        </div>
      </v-sheet>
    </v-dialog>

    <!-- Wait dialog for assign/unassign -->
    <v-dialog v-model="waitAssignUnassignMemory">
      <v-row align-content="center" class="fill-height" justify="center">
        <v-col cols="6">
          <v-progress-linear color="#1428A0" height="50" indeterminate rounded>
            <template v-slot:default>
              <div v-if="assign">{{ assignMemoryProgressText }}</div>
              <div v-else>{{ unassignMemoryProgressText }}</div>
            </template>
          </v-progress-linear>
        </v-col>
      </v-row>
    </v-dialog>

    <v-dialog v-model="assignUnassignSuccess" max-width="600px">
      <v-sheet
        elevation="12"
        max-width="600"
        rounded="lg"
        width="100%"
        class="pa-4 text-center mx-auto"
      >
        <v-icon
          class="mb-5"
          color="success"
          icon="mdi-check-circle"
          size="112"
        ></v-icon>
        <h2 class="text-h5 mb-6">{{ this.operation }} memory succeeded!</h2>
        <p class="mb-4 text-medium-emphasis text-body-2">
          Memory ID:
          <br />{{ this.selectedMemoryRegion?.id }}
        </p>
        <v-divider class="mb-4"></v-divider>
        <div class="text-end">
          <v-btn
            class="text-none"
            color="success"
            rounded
            variant="flat"
            width="90"
            id="assignOrUnassignSuccess"
            @click="assignUnassignSuccess = false"
          >
            Done
          </v-btn>
        </div>
      </v-sheet>
    </v-dialog>

    <v-dialog v-model="assignUnassignFailure" max-width="600px">
      <v-sheet
        elevation="12"
        max-width="600"
        rounded="lg"
        width="100%"
        class="pa-4 text-center mx-auto"
      >
        <v-icon
          class="mb-5"
          color="error"
          icon="mdi-alert-circle"
          size="112"
        ></v-icon>
        <h2 class="text-h5 mb-6">{{ this.operation }} memory failed!</h2>
        <p class="mb-4 text-medium-emphasis text-body-2">
          {{ this.assignUnassignMemoryError }}
        </p>
        <v-divider class="mb-4"></v-divider>
        <div class="text-end">
          <v-btn
            class="text-none"
            color="error"
            rounded
            variant="flat"
            width="90"
            id="assignOrUnassignFailure"
            @click="assignUnassignFailure = false"
          >
            Done
          </v-btn>
        </div>
      </v-sheet>
    </v-dialog>

    <!-- Shared: choose which associated USP port to free (unassign+free combined) -->
    <v-dialog v-model="dialogSelectUspForFree" max-width="600px">
      <v-card>
        <v-alert color="warning" icon="$warning" title="Alert" variant="tonal">
          This Resource is <strong>Shared</strong>. Select which Upstream Port (USP) association to remove.
        </v-alert>
        <v-divider></v-divider>
        <v-card-text>
          <div>
            ResourceId: <strong>{{ selectedUnusedResource?.id }}</strong>
          </div>
          <v-autocomplete
            v-model="selectedUspPortForFree"
            label="Target USP Port"
            :items="associatedUspPortsForSelectedResource"
            id="inputSelectUspForFree"
          ></v-autocomplete>
        </v-card-text>
        <v-divider></v-divider>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn
            color="yellow-darken-4"
            variant="text"
            id="cancelSelectUspForFree"
            @click="dialogSelectUspForFree = false"
          >Cancel</v-btn>
          <v-btn
            color="#1428A0"
            variant="text"
            :disabled="!selectedUspPortForFree"
            id="confirmSelectUspForFree"
            @click="confirmSelectUspForFree"
          >Continue</v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>

    <v-dialog v-model="dialogFreeMemory" max-width="600px">
      <v-card>
        <v-alert
          color="warning"
          icon="$warning"
          title="Alert"
          variant="tonal"
          text="Due to limited protections, the CXL-Host MUST be powered down when being unassigned memory from a switch."
        ></v-alert>
        <v-card-text>
          Please <strong>power down </strong>the connected CXL-Host device
          before clicking the <strong>FREE</strong> button to free the memory:
          <strong>{{ this.selectedMemoryRegion?.id }}</strong>
        </v-card-text>
        <v-divider></v-divider>
        <v-card-actions>
          <v-spacer></v-spacer>
          <v-btn
            color="yellow-darken-4"
            variant="text"
            id="cancelFreeMemory"
            @click="dialogFreeMemory = false"
          >Cancel</v-btn
          >
          <v-btn
            color="#1428A0"
            variant="text"
            id="confirmFreeMemory"
            @click="freeMemoryRegionConfirm"
          >Free</v-btn
          >
        </v-card-actions>
      </v-card>
    </v-dialog>

    <v-dialog v-model="dialogFreeMemoryWait">
      <v-row align-content="center" class="fill-height" justify="center">
        <v-col cols="6">
          <v-progress-linear color="#1428A0" height="50" indeterminate rounded>
            <template v-slot:default>
              <div class="text-center">
                {{ freeMemoryProgressText }}
              </div>
            </template>
          </v-progress-linear>
        </v-col>
      </v-row>
    </v-dialog>

    <v-dialog v-model="dialogFreeMemorySuccess" max-width="600px">
      <v-sheet
        elevation="12"
        max-width="600"
        rounded="lg"
        width="100%"
        class="pa-4 text-center mx-auto"
      >
        <v-icon
          class="mb-5"
          color="success"
          icon="mdi-check-circle"
          size="112"
        ></v-icon>
        <h2 class="text-h5 mb-6">Free memory succeeded!</h2>
        <p class="mb-4 text-medium-emphasis text-body-2">
          Memory ID:
          <br />{{ this.selectedMemoryRegion?.id }}
        </p>
        <v-divider class="mb-4"></v-divider>
        <div class="text-end">
          <v-btn
            class="text-none"
            color="success"
            rounded
            variant="flat"
            width="90"
            id="freeMemorySuccess"
            @click="dialogFreeMemorySuccess = false"
          >
            Done
          </v-btn>
        </div>
      </v-sheet>
    </v-dialog>

    <v-dialog v-model="dialogFreeMemoryFailure" max-width="600px">
      <v-sheet
        elevation="12"
        max-width="600"
        rounded="lg"
        width="100%"
        class="pa-4 text-center mx-auto"
      >
        <v-icon
          class="mb-5"
          color="error"
          icon="mdi-alert-circle"
          size="112"
        ></v-icon>
        <h2 class="text-h5 mb-6">Free memory failed!</h2>
        <p class="mb-4 text-medium-emphasis text-body-2">
          {{ this.freeMemoryError }}
        </p>
        <v-divider class="mb-4"></v-divider>
        <div class="text-end">
          <v-btn
            class="text-none"
            color="error"
            rounded
            variant="flat"
            width="90"
            id="freeMemoryFailure"
            @click="dialogFreeMemoryFailure = false"
          >
            Done
          </v-btn>
        </div>
      </v-sheet>
    </v-dialog>
  </v-container>
</template>

<script>
import { getColor } from "../Common/helpers";
import { computed, ref } from "vue";
import { useBladeResourceStore } from "../Stores/BladeResourceStore";
import { useBladeMemoryStore } from "../Stores/BladeMemoryStore";
import { useBladePortStore } from "../Stores/BladePortStore";
import { useBladeStore } from "../Stores/BladeStore";
import { useApplianceStore } from "../Stores/ApplianceStore";

export default {
  computed: {
    // Assign-to-Port list for the currently selected Resource (Assign dialog)
    // Rules:
    // - Unused   (USP count = 0): list ALL USP ports
    // - Composed (USP count = 1): list ALL USP ports EXCEPT the existing USP port
    // - Shared   (USP count >=2): list ALL USP ports EXCEPT the USP ports already associated
    assignableUspPortsForSelectedResource() {
      const bladePortStore = useBladePortStore();
      const uspPorts = (bladePortStore.bladePorts || [])
        .filter((p) => (p.portType || "").toString().toUpperCase() === "USP")
        .map((p) => p.id);

      const rid = this.selectedUnusedResource?.id;
      if (!rid) return uspPorts;

      const used = new Set(this.getAssociatedUspPortIdsForResourceId(rid));
      const status = this.getDerivedCompositionStateByResourceId(rid);

      if (status === "UNUSED") return uspPorts;
      // Composed/Shared: exclude ports already associated with this resource
      return uspPorts.filter((id) => !used.has(id));
    },

    // Shared Free: select which USP port to free/unassign from
    associatedUspPortsForSelectedResource() {
      const rid = this.selectedUnusedResource?.id;
      if (!rid) return [];
      return this.getAssociatedUspPortIdsForResourceId(rid);
    },
  },

  data() {
    return {
      assignMemoryProgressText: "Assigning Memory, please wait...",
      unassignMemoryProgressText: "Unassigning Memory, please wait...",
      freeMemoryProgressText: "Freeing Memory, please wait...",

      headers: [
        { title: "Status", align: "start", key: "compositionStatus.compositionState" },
        { title: "ResourceId", key: "id"},
        { title: "PPBId", key: "channelId" },
        { title: "LDId", key: "channelResourceIndex" },
        { title: "CapacityGiB", key: "capacityMiB" },
        { title: "Actions", key: "actions", sortable: false },
      ],

      selectedMemoryRegion: null,

      waitAssignUnassignMemory: false,

      dialogAssignUnassign: false,
      assign: false,

      assignUnassignMemory: {
        port: "",
        operation: "",
      },
      operation: "",

      assignUnassignSuccess: false,
      assignUnassignFailure: false,
      assignUnassignMemoryError: null,
      assignOrUnassignResponse: null,
      dialogNoPortForAssignMemory: false,

      dialogFreeMemory: false,
      dialogFreeMemoryWait: false,
      freeMemoryError: null,
      freeMemoryResponse: null,
      dialogFreeMemorySuccess: false,
      dialogFreeMemoryFailure: false,

      // v-autocomplete model
      assignPort: "",

      // Unused Resource -> Assign by selecting an Upstream Port
      dialogAssignUnusedResource: false,
      waitAssignUnusedResource: false,
      assignUnusedResourceSuccess: false,
      assignUnusedResourceFailure: false,
      assignUnusedResourceError: null,
      selectedUnusedResource: null,
      assignUnusedResourcePort: "",
      assignedUnusedResourcePortSnapshot: "",
      newComposedMemoryId: "",

      // Shared Free: choose which USP port to remove
      dialogSelectUspForFree: false,
      selectedUspPortForFree: "",
    };
  },

  methods: {
    getStatusColor(item) {
      return getColor(item);
    },

    normalizeCompositionState(resourceBlock) {
      // Status is derived from the number of associated USP ports (request requirement)
      // - Unused   : 0 USP ports
      // - Composed : 1 USP port
      // - Shared   : >=2 USP ports
      if (resourceBlock?.id) {
        return this.getDerivedCompositionStateByResourceId(resourceBlock.id);
      }
      return ((resourceBlock?.compositionStatus?.compositionState || "").toString().trim().toUpperCase());
    },

    // Return the list of USP port IDs that are currently associated with a ResourceId.
    // Association is inferred by scanning BladeMemoryStore memory regions that reference the ResourceId,
    // and keeping only those whose memoryAppliancePort is a USP port.
    getAssociatedUspPortIdsForResourceId(resourceId) {
      const bladeMemoryStore = useBladeMemoryStore();
      const bladePortStore = useBladePortStore();

      const uspPortIdSet = new Set(
        (bladePortStore.bladePorts || [])
          .filter((p) => (p.portType || "").toString().toUpperCase() === "USP")
          .map((p) => p.id)
      );

      const portIds = (bladeMemoryStore.bladeMemory || [])
        .filter((mr) => (mr.resourceIds || []).includes(resourceId))
        .map((mr) => mr.memoryAppliancePort)
        .filter((pid) => !!pid && uspPortIdSet.has(pid));

      // unique, stable order
      return Array.from(new Set(portIds)).sort();
    },

    // Derived status by ResourceId (see normalizeCompositionState)
    getDerivedCompositionStateByResourceId(resourceId) {
      const uspPorts = this.getAssociatedUspPortIdsForResourceId(resourceId);
      const n = uspPorts.length;
      if (n === 0) return "UNUSED";
      if (n === 1) return "COMPOSED";
      return "SHARED";
    },

    isComposed(resourceBlock) {
      return this.normalizeCompositionState(resourceBlock) === "COMPOSED";
    },

    isShared(resourceBlock) {
      return this.normalizeCompositionState(resourceBlock) === "SHARED";
    },

    // ✅ ResourceId -> MemoryRegion mapping
    findMemoryRegionByResourceId(resourceId) {
      const bladeMemoryStore = useBladeMemoryStore();
      return bladeMemoryStore.bladeMemory.find((mr) =>
        (mr.resourceIds || []).includes(resourceId)
      );
    },

    // Unused resource detection
    isUnused(resourceBlock) {
      return (
        this.normalizeCompositionState(resourceBlock) === "UNUSED"
      );
    },

    // Assign entry point depending on current status
    openAssignDialogByStatus(resourceBlock) {
      // Single Assign dialog for all statuses.
      // Port list is computed in assignableUspPortsForSelectedResource.
      this.openAssignUnusedResourceDialog(resourceBlock);

      // For Composed/Shared: if no eligible ports (all already associated), show "no port" dialog.
      if (!this.isUnused(resourceBlock)) {
        const eligible = this.assignableUspPortsForSelectedResource;
        if (!eligible || eligible.length === 0) {
          this.dialogAssignUnusedResource = false;
          this.dialogNoPortForAssignMemory = true;
        }
      }
    },

    // Local UI helper: update composition state for one or more resources
    updateLocalResourceStates(resourceIds, newState) {
      const bladeResourceStore = useBladeResourceStore();
      (resourceIds || []).forEach((rid) => {
        const localRes = bladeResourceStore.memoryResources.find((r) => r.id === rid);
        if (localRes?.compositionStatus) {
          localRes.compositionStatus.compositionState = newState;
        }
      });
    },

    // After assign/free operations, refresh-derived status (Unused/Composed/Shared)
    // based on current BladeMemoryStore + BladePortStore associations.
    syncDerivedStateForResourceIds(resourceIds) {
      const ids = (resourceIds || []).filter(Boolean);
      if (ids.length === 0) return;
      const bladeResourceStore = useBladeResourceStore();
      ids.forEach((rid) => {
        const derived = this.getDerivedCompositionStateByResourceId(rid);
        const localRes = bladeResourceStore.memoryResources.find((r) => r.id === rid);
        if (localRes?.compositionStatus) {
          if (derived === "UNUSED") localRes.compositionStatus.compositionState = "Unused";
          else if (derived === "COMPOSED") localRes.compositionStatus.compositionState = "Composed";
          else localRes.compositionStatus.compositionState = "Shared";
        }
      });
    },

    openAssignUnusedResourceDialog(resourceBlock) {
      this.selectedUnusedResource = resourceBlock;
      this.assignUnusedResourcePort = "";
      this.assignedUnusedResourcePortSnapshot = "";
      this.newComposedMemoryId = "";
      this.assignUnusedResourceError = null;
      this.dialogAssignUnusedResource = true;
    },

    async assignUnusedResourceConfirm() {
      const applianceStore = useApplianceStore();
      const bladeStore = useBladeStore();
      const bladeMemoryStore = useBladeMemoryStore();
      const bladeResourceStore = useBladeResourceStore();
      const bladePortStore = useBladePortStore();

      this.dialogAssignUnusedResource = false;

      if (!this.selectedUnusedResource) {
        this.assignUnusedResourceError = "No Resource selected.";
        this.assignUnusedResourceFailure = true;
        return;
      }
      if (!this.assignUnusedResourcePort) {
        this.assignUnusedResourceError = "Assign port need to be selected.";
        this.assignUnusedResourceFailure = true;
        return;
      }
      if (!applianceStore.selectedApplianceId || !bladeStore.selectedBladeId) {
        this.assignUnusedResourceError = "No selected appliance/blade.";
        this.assignUnusedResourceFailure = true;
        return;
      }

      // Capture status before mutation (derived by USP-port count)
      const prevStatus = this.getDerivedCompositionStateByResourceId(this.selectedUnusedResource.id);

      this.waitAssignUnusedResource = true;
      this.assignedUnusedResourcePortSnapshot = this.assignUnusedResourcePort;

      const composeReq = {
        Port: this.assignUnusedResourcePort,
        memoryResources: [this.selectedUnusedResource.id],
      };

      const newMemory = await bladeMemoryStore.composeMemoryByResource(
        applianceStore.selectedApplianceId,
        bladeStore.selectedBladeId,
        composeReq
      );

      if (newMemory) {
        this.newComposedMemoryId = newMemory.id || "";

        // -----------------------------------------------------
        // UI sync fixes
        // 1) Resources.vue Status: force the selected resource to "Composed"
        // 2) Memory.vue Allocated/Available: update BladeStore totals
        // -----------------------------------------------------

        // (1) Optimistic local update per requirement:
        // - Unused   -> Assign => Composed
        // - Composed -> Assign => Shared
        // - Shared   -> Assign => Shared
        const nextState = prevStatus === "UNUSED" ? "Composed" : "Shared";
        this.updateLocalResourceStates([this.selectedUnusedResource?.id], nextState);

        // (2) Update blade memory totals (backend values are MiB; UI memory region size is GiB)
        const deltaMiB = (newMemory.sizeMiB || 0) * 1024;
        if (deltaMiB > 0) {
          const curAvail = bladeStore.selectedBladeTotalMemoryAvailableMiB;
          const curAlloc = bladeStore.selectedBladeTotalMemoryAllocatedMiB;

          // Be conservative with undefined/0 handling
          const newAvailable =
            typeof curAvail === "number" ? curAvail - deltaMiB : undefined;
          const newAllocated =
            typeof curAlloc === "number" ? curAlloc + deltaMiB : deltaMiB;

          bladeStore.updateSelectedBladeMemory(newAvailable, newAllocated);
        }

        // Refresh related data (so derived status based on USP associations becomes correct)
        await bladeMemoryStore.fetchBladeMemory(
          applianceStore.selectedApplianceId,
          bladeStore.selectedBladeId
        );

        // ResourceStatus endpoint can lag; refetch resources too to keep UI consistent
        await bladeResourceStore.fetchMemoryResources(
          applianceStore.selectedApplianceId,
          bladeStore.selectedBladeId
        );
        await bladeResourceStore.updateMemoryResourcesStatus(
          applianceStore.selectedApplianceId,
          bladeStore.selectedBladeId
        );
        await bladePortStore.fetchBladePorts(
          applianceStore.selectedApplianceId,
          bladeStore.selectedBladeId
        );

        // Ensure Resources table status matches the derived rule after refresh
        this.syncDerivedStateForResourceIds([this.selectedUnusedResource?.id]);

        this.waitAssignUnusedResource = false;
        this.assignUnusedResourceSuccess = true;
      } else {
        this.assignUnusedResourceError =
          bladeMemoryStore.composeMemoryByResourceError ||
          "Compose/assign by resource failed.";
        this.waitAssignUnusedResource = false;
        this.assignUnusedResourceFailure = true;
      }
    },

    // ✅ wrappers from Resources row
    assignOrUnassignFromResource(resourceBlock) {
      const mr = this.findMemoryRegionByResourceId(resourceBlock.id);
      if (!mr) return; // conservative: do nothing if mapping fails
      this.assignOrUnassign(mr);
    },

    freeMemoryFromResource(resourceBlock) {
      const rid = resourceBlock?.id;
      if (!rid) return;

      // Shared: let user choose which associated USP port to free
      if (this.isShared(resourceBlock)) {
        this.selectedUnusedResource = resourceBlock;
        this.selectedUspPortForFree = "";
        this.dialogSelectUspForFree = true;
        return;
      }

      // Composed: only one associated USP, free the corresponding memory region
      const mr = this.findMemoryRegionByResourceId(rid);
      if (!mr) return;
      this.freeMemory(mr);
    },

    confirmSelectUspForFree() {
      const rid = this.selectedUnusedResource?.id;
      const portId = this.selectedUspPortForFree;
      if (!rid || !portId) {
        this.dialogSelectUspForFree = false;
        return;
      }

      const bladeMemoryStore = useBladeMemoryStore();
      const mr = (bladeMemoryStore.bladeMemory || []).find(
        (m) => (m.resourceIds || []).includes(rid) && m.memoryAppliancePort === portId
      );

      this.dialogSelectUspForFree = false;
      this.selectedUspPortForFree = "";

      if (!mr) return;
      this.freeMemory(mr);
    },

    // =========================
    // Below methods are copied from Memory.vue (as-is, operating on MemoryRegion)
    // =========================

    assignOrUnassign(item) {
      this.selectedMemoryRegion = item;

      if (this.selectedMemoryRegion.memoryAppliancePort) {
        this.operation = "unassign";
        this.assign = false;
        this.dialogAssignUnassign = true;
      } else {
        if (this.LengthOfPorts == 0) {
          this.dialogNoPortForAssignMemory = true;
        } else {
          this.operation = "assign";
          this.assign = true;
          this.dialogAssignUnassign = true;
        }
      }
    },

    async assignUnassignPort(operation) {
      const bladeMemoryStore = useBladeMemoryStore();

      this.dialogAssignUnassign = false;

      if (this.selectedMemoryRegion) {
        this.waitAssignUnassignMemory = true;

        if (this.selectedMemoryRegion.memoryAppliancePort) {
          this.assignUnassignMemory.port =
            this.selectedMemoryRegion.memoryAppliancePort;
        } else {
          if (this.assignPort) {
            this.assignUnassignMemory.port = this.assignPort;
            this.assignPort = "";
          } else {
            this.assignUnassignMemoryError = "Assign port need to be selected.";
            this.waitAssignUnassignMemory = false;
            this.assignUnassignFailure = true;
            return;
          }
        }

        this.assignUnassignMemory.operation = operation;

        this.assignOrUnassignResponse = await bladeMemoryStore.assignOrUnassign(
          this.selectedMemoryRegion.memoryApplianceId,
          this.selectedMemoryRegion.memoryBladeId,
          this.selectedMemoryRegion.id,
          this.assignUnassignMemory
        );
        this.assignUnassignMemoryError =
          bladeMemoryStore.assignOrUnassignMemoryError;

        if (this.assignOrUnassignResponse) {
          const bladeResourceStore = useBladeResourceStore();
          const bladePortStore = useBladePortStore();

          // Optimistic UI update for Resources.vue
          // - unassign => Unused
          // - assign   => Shared
          if (operation === "unassign") {
            this.updateLocalResourceStates(this.selectedMemoryRegion?.resourceIds, "Unused");
          } else if (operation === "assign") {
            this.updateLocalResourceStates(this.selectedMemoryRegion?.resourceIds, "Shared");
          }

          await bladeMemoryStore.fetchBladeMemory(
            this.selectedMemoryRegion.memoryApplianceId,
            this.selectedMemoryRegion.memoryBladeId
          );
          await bladeResourceStore.updateMemoryResourcesStatus(
            this.selectedMemoryRegion.memoryApplianceId,
            this.selectedMemoryRegion.memoryBladeId
          );
          await bladePortStore.fetchBladePorts(
            this.selectedMemoryRegion.memoryApplianceId,
            this.selectedMemoryRegion.memoryBladeId
          );

          this.waitAssignUnassignMemory = false;
          this.assignUnassignSuccess = true;
        } else {
          this.waitAssignUnassignMemory = false;
          this.assignUnassignFailure = true;
        }
      }
    },

    freeMemory(item) {
      this.selectedMemoryRegion = item;
      this.dialogFreeMemory = true;
    },

    async freeMemoryRegionConfirm() {
      this.dialogFreeMemory = false;
      this.dialogFreeMemoryWait = true;

      const bladeMemoryStore = useBladeMemoryStore();
      if (this.selectedMemoryRegion) {
        this.freeMemoryResponse = await bladeMemoryStore.freeMemory(
          this.selectedMemoryRegion.memoryApplianceId,
          this.selectedMemoryRegion.memoryBladeId,
          this.selectedMemoryRegion.id
        );
        this.freeMemoryError = bladeMemoryStore.freeMemoryError;
      }

      if (!this.freeMemoryError) {
        const bladeResourceStore = useBladeResourceStore();
        const bladePortStore = useBladePortStore();

        const affectedResourceIds = this.selectedMemoryRegion?.resourceIds || [];

        // Refresh so status can be derived from USP-port associations (Unused/Composed/Shared)
        await bladeMemoryStore.fetchBladeMemory(
          this.selectedMemoryRegion.memoryApplianceId,
          this.selectedMemoryRegion.memoryBladeId
        );
        await bladeResourceStore.fetchMemoryResources(
          this.selectedMemoryRegion.memoryApplianceId,
          this.selectedMemoryRegion.memoryBladeId
        );
        await bladeResourceStore.updateMemoryResourcesStatus(
          this.selectedMemoryRegion.memoryApplianceId,
          this.selectedMemoryRegion.memoryBladeId
        );
        await bladePortStore.fetchBladePorts(
          this.selectedMemoryRegion.memoryApplianceId,
          this.selectedMemoryRegion.memoryBladeId
        );

        // Ensure displayed Status matches the derived rule
        this.syncDerivedStateForResourceIds(affectedResourceIds);

        const bladeStore = useBladeStore();
        let newAvailableMemory;
        let newAllocatedMmeory;

        if (
          bladeStore.selectedBladeTotalMemoryAvailableMiB &&
          bladeStore.selectedBladeTotalMemoryAllocatedMiB
        ) {
          newAvailableMemory =
            bladeStore.selectedBladeTotalMemoryAvailableMiB +
            this.selectedMemoryRegion.sizeMiB * 1024;
          newAllocatedMmeory =
            bladeStore.selectedBladeTotalMemoryAllocatedMiB -
            this.selectedMemoryRegion.sizeMiB * 1024;
        } else if (bladeStore.selectedBladeTotalMemoryAllocatedMiB) {
          newAvailableMemory = this.selectedMemoryRegion.sizeMiB * 1024;
          newAllocatedMmeory =
            bladeStore.selectedBladeTotalMemoryAllocatedMiB -
            this.selectedMemoryRegion.sizeMiB * 1024;
        }

        await bladeStore.updateSelectedBladeMemory(
          newAvailableMemory,
          newAllocatedMmeory
        );

        this.dialogFreeMemoryWait = false;
        this.dialogFreeMemorySuccess = true;
      } else {
        this.dialogFreeMemoryWait = false;
        this.dialogFreeMemoryFailure = true;
      }
    },
  },

  setup() {
    const bladeResourceStore = useBladeResourceStore();
    const bladeMemoryStore = useBladeMemoryStore();
    const bladePortStore = useBladePortStore();

    const activeResourceIdFilter = computed(() => bladeResourceStore.resourceIdFilter);

    // Sort resources by numeric part of ResourceId, and apply optional filter
    const sortedBladeResources = computed(() => {
      const filtered = bladeResourceStore.resourceIdFilter
        ? bladeResourceStore.memoryResources.filter(
          (r) => r.id === bladeResourceStore.resourceIdFilter
        )
        : bladeResourceStore.memoryResources;

      return filtered
        .slice()
        .sort((a, b) => {
          const numA = parseInt(a.id.replace(/^\D+/g, ""));
          const numB = parseInt(b.id.replace(/^\D+/g, ""));
          return numA - numB;
        });
    });

    // Status is derived from the number of associated USP ports (0/1/2+)
    const derivedBladeResources = computed(() => {
      const uspPortIdSet = new Set(
        (bladePortStore.bladePorts || [])
          .filter((p) => (p.portType || "").toString().toUpperCase() === "USP")
          .map((p) => p.id)
      );

      const deriveStateTitle = (resourceId) => {
        const ports = (bladeMemoryStore.bladeMemory || [])
          .filter((mr) => (mr.resourceIds || []).includes(resourceId))
          .map((mr) => mr.memoryAppliancePort)
          .filter((pid) => !!pid && uspPortIdSet.has(pid));
        const n = new Set(ports).size;
        if (n === 0) return "Unused";
        if (n === 1) return "Composed";
        return "Shared";
      };

      return (sortedBladeResources.value || []).map((r) => {
        const derivedTitle = deriveStateTitle(r.id);
        return {
          ...r,
          compositionStatus: {
            ...(r.compositionStatus || {}),
            compositionState: derivedTitle,
          },
        };
      });
    });

    // Port lists
    // - Only USP ports are eligible for Assign actions in Resources.vue
    const uspPortIds = computed(() =>
      bladePortStore.bladePorts
        .filter((port) => (port.portType || "").toString().toUpperCase() === "USP")
        .map((port) => port.id)
    );

    const assignedPortIds = computed(() =>
      bladeMemoryStore.bladeMemory
        .filter((memoryRegion) => memoryRegion.memoryAppliancePort)
        .map((memoryRegion) => memoryRegion.memoryAppliancePort)
    );

    const unassignedUspPortIds = computed(() => {
      const assignedIdsSet = new Set(assignedPortIds.value);
      return uspPortIds.value.filter((id) => !assignedIdsSet.has(id));
    });

    const lengthOfPorts = computed(() => {
      return unassignedUspPortIds.value.length;
    });

    return {
      selectedBladeResources: derivedBladeResources,
      PortIdArray: unassignedUspPortIds,
      uspPortIds,
      LengthOfPorts: lengthOfPorts,
      activeResourceIdFilter,
      clearResourceFilter: () => bladeResourceStore.setResourceIdFilter(null),
      unassignedUspPortIds,
    };
  },
};
</script>
