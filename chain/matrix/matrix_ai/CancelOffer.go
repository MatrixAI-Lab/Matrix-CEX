// Code generated by https://github.com/gagliardetto/anchor-go. DO NOT EDIT.

package matrix_ai

import (
	"errors"
	ag_binary "github.com/gagliardetto/binary"
	ag_solanago "github.com/gagliardetto/solana-go"
	ag_format "github.com/gagliardetto/solana-go/text/format"
	ag_treeout "github.com/gagliardetto/treeout"
)

// CancelOffer is the `cancelOffer` instruction.
type CancelOffer struct {

	// [0] = [WRITE] machine
	//
	// [1] = [WRITE, SIGNER] owner
	ag_solanago.AccountMetaSlice `bin:"-"`
}

// NewCancelOfferInstructionBuilder creates a new `CancelOffer` instruction builder.
func NewCancelOfferInstructionBuilder() *CancelOffer {
	nd := &CancelOffer{
		AccountMetaSlice: make(ag_solanago.AccountMetaSlice, 2),
	}
	return nd
}

// SetMachineAccount sets the "machine" account.
func (inst *CancelOffer) SetMachineAccount(machine ag_solanago.PublicKey) *CancelOffer {
	inst.AccountMetaSlice[0] = ag_solanago.Meta(machine).WRITE()
	return inst
}

// GetMachineAccount gets the "machine" account.
func (inst *CancelOffer) GetMachineAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice.Get(0)
}

// SetOwnerAccount sets the "owner" account.
func (inst *CancelOffer) SetOwnerAccount(owner ag_solanago.PublicKey) *CancelOffer {
	inst.AccountMetaSlice[1] = ag_solanago.Meta(owner).WRITE().SIGNER()
	return inst
}

// GetOwnerAccount gets the "owner" account.
func (inst *CancelOffer) GetOwnerAccount() *ag_solanago.AccountMeta {
	return inst.AccountMetaSlice.Get(1)
}

func (inst CancelOffer) Build() *Instruction {
	return &Instruction{BaseVariant: ag_binary.BaseVariant{
		Impl:   inst,
		TypeID: Instruction_CancelOffer,
	}}
}

// ValidateAndBuild validates the instruction parameters and accounts;
// if there is a validation error, it returns the error.
// Otherwise, it builds and returns the instruction.
func (inst CancelOffer) ValidateAndBuild() (*Instruction, error) {
	if err := inst.Validate(); err != nil {
		return nil, err
	}
	return inst.Build(), nil
}

func (inst *CancelOffer) Validate() error {
	// Check whether all (required) accounts are set:
	{
		if inst.AccountMetaSlice[0] == nil {
			return errors.New("accounts.Machine is not set")
		}
		if inst.AccountMetaSlice[1] == nil {
			return errors.New("accounts.Owner is not set")
		}
	}
	return nil
}

func (inst *CancelOffer) EncodeToTree(parent ag_treeout.Branches) {
	parent.Child(ag_format.Program(ProgramName, ProgramID)).
		//
		ParentFunc(func(programBranch ag_treeout.Branches) {
			programBranch.Child(ag_format.Instruction("CancelOffer")).
				//
				ParentFunc(func(instructionBranch ag_treeout.Branches) {

					// Parameters of the instruction:
					instructionBranch.Child("Params[len=0]").ParentFunc(func(paramsBranch ag_treeout.Branches) {})

					// Accounts of the instruction:
					instructionBranch.Child("Accounts[len=2]").ParentFunc(func(accountsBranch ag_treeout.Branches) {
						accountsBranch.Child(ag_format.Meta("machine", inst.AccountMetaSlice.Get(0)))
						accountsBranch.Child(ag_format.Meta("  owner", inst.AccountMetaSlice.Get(1)))
					})
				})
		})
}

func (obj CancelOffer) MarshalWithEncoder(encoder *ag_binary.Encoder) (err error) {
	return nil
}
func (obj *CancelOffer) UnmarshalWithDecoder(decoder *ag_binary.Decoder) (err error) {
	return nil
}

// NewCancelOfferInstruction declares a new CancelOffer instruction with the provided parameters and accounts.
func NewCancelOfferInstruction(
	// Accounts:
	machine ag_solanago.PublicKey,
	owner ag_solanago.PublicKey) *CancelOffer {
	return NewCancelOfferInstructionBuilder().
		SetMachineAccount(machine).
		SetOwnerAccount(owner)
}
