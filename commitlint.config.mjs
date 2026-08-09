export default {
	extends: ["@commitlint/config-conventional"],
	parserPreset: {
		name: "conventional-changelog-conventionalcommits",
		presetConfig: {
			types: [
				{ type: "feat", section: "Features" },
				{ type: "fix", section: "Bug Fixes" },
				{ type: "docs", section: "Documentation", hidden: false },
			],
		},
	},
};
